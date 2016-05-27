package dnsserver

import (
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func getResult(ctx context.Context, ticker *time.Ticker, respCh <-chan *dns.Msg) (resp *dns.Msg, responded, canceled bool) {
	select {
	case <-ctx.Done():
		canceled = true
	case resp = <-respCh:
		responded = true
	case <-ticker.C:
	}

	return
}

func (d *DNSServer) fastLookup(ctx context.Context, net string, addr []string, req *dns.Msg) *dns.Msg {
	respCh := make(chan *dns.Msg)

	ticker := time.NewTicker(d.LookupInterval)
	defer ticker.Stop()
	var nresponses int

	// start lookup on each nameserver top-down, every LookupInterval
	for _, nameserver := range addr {
		go d.lookup(net, nameserver, req, respCh)

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

func (d *DNSServer) lookup(net, nameserver string, req *dns.Msg, respCh chan<- *dns.Msg) {
	c := &dns.Client{
		Net:          net,
		DialTimeout:  d.DialTimeout,
		ReadTimeout:  d.ClientTimeout,
		WriteTimeout: d.ClientTimeout,
	}

	sendResponse := func(resp *dns.Msg) {
		select {
		case respCh <- resp:
		default:
		}
	}

	ctxLog := d.Logger.WithFields(logrus.Fields{
		"name":       req.Question[0].Name,
		"type":       dns.TypeToString[req.Question[0].Qtype],
		"net":        net,
		"nameserver": nameserver,
	})

	// TODO(jrubin)
	// Exchange does not retry a failed query, nor will it fall back to TCP in
	// case of truncation.
	// It is up to the caller to create a message that allows for larger responses to be
	// returned. Specifically this means adding an EDNS0 OPT RR that will advertise a larger
	// buffer, see SetEdns0. Messsages without an OPT RR will fallback to the historic limit
	// of 512 bytes.
	resp, _, err := c.Exchange(req, nameserver)
	if err != nil {
		ctxLog.WithError(err).Warn("socket error")
		sendResponse(nil)
		return
	}

	if resp != nil && resp.Rcode != dns.RcodeSuccess && resp.Rcode != dns.RcodeNameError {
		ctxLog.Warn("failed to get a valid answer")
		if resp.Rcode == dns.RcodeServerFailure {
			sendResponse(nil)
			return
		}
	}

	sendResponse(resp)
}
