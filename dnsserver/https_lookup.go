package dnsserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"

	"github.com/certifi/gocertifi"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"jrubin.io/slog"
)

var googleDNSAddrs = []string{
	"8.8.8.8:53",
	"8.8.4.4:53",
	"[2001:4860:4860::8888]:53",
	"[2001:4860:4860::8844]:53",
}

type transport struct {
	transport *http.Transport
	urls      []string
}

type DNSHTTP struct {
	KeepAlive             time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	NoDNSSEC              bool
	EDNSClientSubnet      string
	NoRandomPadding       bool
	transport             map[string]*transport
}

type HTTPSQuestion struct {
	Name string // FQDN with trailing dot
	Type uint16 // Standard DNS RR type
}

type HTTPSRR struct {
	Name string // Always matches name in the Question section
	Type uint16 // Standard DNS RR type
	TTL  uint32 // Record's time-to-live in seconds

	// Data for A - IP address as text
	// Data for SPF - quoted string
	// Data for TXT - multiple quoted strings
	Data string
}

type HTTPSResponse struct {
	Status           int  // Standard DNS response code (32 bit integer)
	TC               bool // Whether the response is truncated
	RD               bool // Always true for Google Public DNS
	RA               bool // Always true for Google Public DNS
	AD               bool // Whether all response data was validated with DNSSEC
	CD               bool // Whether the client asked to disable DNSSEC
	Question         []HTTPSQuestion
	Answer           []HTTPSRR
	Authority        []HTTPSRR
	Additional       []HTTPSRR
	EDNSClientSubnet string `json:"edns_client_subnet"` // IP address / scope prefix-length
	Comment          string
}

func (d *DNSServer) initHTTPTransport(addr string) error {
	if d.HTTP.transport == nil {
		d.HTTP.transport = map[string]*transport{}
	}

	if _, ok := d.HTTP.transport[addr]; ok {
		return nil
	}

	rootCAs, err := gocertifi.CACerts()
	if err != nil {
		return err
	}

	u, err := url.Parse(addr)
	if err != nil {
		return err
	}

	var host, port string

	if host, port, err = net.SplitHostPort(u.Host); err != nil {
		host = u.Host
	}

	httpTransport := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   d.DialTimeout,
			KeepAlive: d.HTTP.KeepAlive,
		}).DialContext,
		MaxIdleConns:          d.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost:   d.HTTP.MaxIdleConnsPerHost,
		IdleConnTimeout:       d.HTTP.IdleConnTimeout,
		TLSHandshakeTimeout:   d.HTTP.TLSHandshakeTimeout,
		ResponseHeaderTimeout: d.ClientTimeout,
		ExpectContinueTimeout: d.HTTP.ExpectContinueTimeout,
		TLSClientConfig: &tls.Config{
			ServerName: host, // required since we are changing the query url to use an IP
			MinVersion: tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
			RootCAs: rootCAs,
		},
	}

	if err = http2.ConfigureTransport(&httpTransport); err != nil {
		return err
	}

	t := transport{
		transport: &httpTransport,
	}

	if ip := net.ParseIP(host); ip != nil {
		t.urls = append(t.urls, u.String())
		d.HTTP.transport[addr] = &t
		return nil
	}

	// lookup the ip address using google public dns (this is the only query
	// that will be exposed)
	ips := d.easyLookup(context.Background(), host+".", googleDNSAddrs...)
	if len(ips) == 0 {
		return errors.New("could not determine ips for https dns host: " + host)
	}

	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			u.Host = ip.String()
		} else if ip6 := ip.To16(); ip6 != nil {
			u.Host = "[" + ip.String() + "]"
		} else {
			return errors.New("invalid ip address: " + ip.String())
		}

		if len(port) > 0 {
			// NOTE: don't use net.JoinHostPort as it may add extra brackets
			u.Host += ":" + port
		}

		t.urls = append(t.urls, u.String())
	}

	d.HTTP.transport[addr] = &t

	return nil
}

const queryLen = 512

func randString(size int) (string, error) {
	if size < 0 {
		return "", nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, size/2+1))

	if _, err := io.Copy(buf, io.LimitReader(rand.Reader, int64(size/2+1))); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes())[:size], nil
}

func (d *DNSServer) fastHTTPSLookup(ctx context.Context, addr string, req *dns.Msg) *dns.Msg {
	respCh := make(chan *dns.Msg)

	ticker := time.NewTicker(d.LookupInterval)
	defer ticker.Stop()
	var nresponses int

	t := d.HTTP.transport[addr]

	// start lookup on each nameserver top-down, every LookupInterval
	for _, nameserver := range t.urls {
		go d.httpsLookup(ctx, t.transport, nameserver, req, respCh)

		// but exit early, if we have an answer
		if resp, responded, canceled := getResult(ctx, ticker, respCh); resp != nil || canceled {
			return resp
		} else if responded {
			nresponses++
		}
	}

	ticker.Stop()

	for i := nresponses; i < len(addr); i++ {
		if resp, _, canceled := getResult(ctx, ticker, respCh); resp != nil || canceled {
			return resp
		}
	}

	return nil
}

func (d *DNSServer) httpsLookup(ctx context.Context, t *http.Transport, urlStr string, req *dns.Msg, respCh chan<- *dns.Msg) {
	ctxLog := d.Logger.WithFields(slog.Fields{
		"name":       req.Question[0].Name,
		"type":       dns.TypeToString[req.Question[0].Qtype],
		"https_host": t.TLSClientConfig.ServerName,
		"https_url":  urlStr,
	})

	sendResponse := func(resp *dns.Msg) {
		select {
		case respCh <- resp:
		default:
		}
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		ctxLog.WithError(err).Warn("error parsing url")
		sendResponse(nil)
		return
	}

	query := u.Query()
	query.Set("name", req.Question[0].Name)
	query.Set("type", fmt.Sprintf("%d", req.Question[0].Qtype))

	if d.HTTP.NoDNSSEC {
		query.Set("cd", "1") // disable dnssec checking
	}

	if len(d.HTTP.EDNSClientSubnet) > 0 {
		query.Set("edns_client_subnet", "")
	}

	if !d.HTTP.NoRandomPadding {
		var str string
		if str, err = randString(queryLen - len(query.Encode())); err != nil {
			ctxLog.WithError(err).Warn("error generating random string")
			sendResponse(nil)
			return
		}

		query.Set("random_padding", str)
	}

	u.RawQuery = query.Encode()

	hreq, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		ctxLog.WithError(err).Warn("error creating http request")
		sendResponse(nil)
		return
	}

	hreq.Host = t.TLSClientConfig.ServerName

	if ctx != nil {
		hreq.WithContext(ctx)
	}

	hresp, err := (&http.Client{Transport: t}).Do(hreq)
	if err != nil {
		ctxLog.WithError(err).Warn("error making https request")
		sendResponse(nil)
		return
	}

	if hresp.StatusCode != http.StatusOK {
		ctxLog.WithFields(slog.Fields{
			"code":   hresp.StatusCode,
			"status": http.StatusText(hresp.StatusCode),
		}).Warn("https request returned unexpected status code")
		sendResponse(nil)
		return
	}

	defer ignoreError(hresp.Body.Close)

	dec := json.NewDecoder(hresp.Body)

	var dresp HTTPSResponse
	if err := dec.Decode(&dresp); err != nil {
		ctxLog.WithError(err).Warn("error decoding https json response")
		sendResponse(nil)
		return
	}

	var resp dns.Msg
	resp.SetReply(req)
	resp.Rcode = dresp.Status
	resp.Truncated = dresp.TC
	resp.RecursionDesired = dresp.RD
	resp.RecursionAvailable = dresp.RA
	resp.AuthenticatedData = dresp.AD
	resp.CheckingDisabled = dresp.CD

	for _, s := range []struct {
		HTTPRRs []HTTPSRR
		DNSRRs  *[]dns.RR
	}{{
		HTTPRRs: dresp.Answer,
		DNSRRs:  &resp.Answer,
	}, {
		HTTPRRs: dresp.Authority,
		DNSRRs:  &resp.Ns,
	}, {
		HTTPRRs: dresp.Additional,
		DNSRRs:  &resp.Extra,
	}} {
		for _, hrr := range s.HTTPRRs {
			rr, err := hrr.DNSRR()
			if err != nil {
				ctxLog.WithError(err).Warn("error parsing https rr")
				continue
			}

			*s.DNSRRs = append(*s.DNSRRs, rr)
		}
	}

	sendResponse(&resp)
}

func (hrr HTTPSRR) DNSRR() (dns.RR, error) {
	hdr := dns.RR_Header{
		Name:   hrr.Name,
		Rrtype: hrr.Type,
		Class:  dns.ClassINET,
		Ttl:    hrr.TTL,
	}

	return dns.NewRR(hdr.String() + hrr.Data)
}

func ignoreError(fn func() error) {
	_ = fn()
}
