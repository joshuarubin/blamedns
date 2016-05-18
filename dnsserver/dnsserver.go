package dnsserver

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

type Blocker interface {
	Block(host string) bool
}

type Passer interface {
	Pass(host string) bool
}

type Block struct {
	IPv4, IPv6 net.IP
	Host       string
	TTL        time.Duration
	Blockers   []Blocker
	Passers    []Passer
}

type DNSServer struct {
	Forward []string
	Servers []*dns.Server
	Block   Block
	Logger  *logrus.Logger
}

const defaultDNSPort = 53

// TODO(jrubin) handle other question types?
var respondToQtypes = []uint16{dns.TypeA, dns.TypeAAAA, dns.TypeCNAME}

func shouldRespondToQtype(q uint16) bool {
	for _, t := range respondToQtypes {
		if t == q {
			return true
		}
	}
	return false
}

func addPort(addrs []string) ([]string, error) {
	ret := make([]string, len(addrs))

	for i, addr := range addrs {
		_, _, err := net.SplitHostPort(addr)
		if err == nil {
			ret[i] = addr
			continue
		}

		if aerr, ok := err.(*net.AddrError); ok {
			if aerr.Err == "missing port in address" {
				ret[i] = fmt.Sprintf("%s:%d", addr, defaultDNSPort)
				continue
			}
		}

		return nil, err
	}

	return ret, nil
}

func (d *DNSServer) ListenAndServe() error {
	var err error
	if d.Forward, err = addPort(d.Forward); err != nil {
		return err
	}

	d.Logger.Infof("using forwarders: %s", strings.Join(d.Forward, ", "))

	errCh := make(chan error, len(d.Servers))

	for _, server := range d.Servers {
		server.Handler = d
		go func(server *dns.Server) {
			d.Logger.WithFields(logrus.Fields{
				"Net":  server.Net,
				"Addr": server.Addr,
			}).Info("starting dns server")
			errCh <- server.ListenAndServe()
		}(server)
	}

	for range d.Servers {
		if err := <-errCh; err != nil {
			return err
		}
	}

	return nil
}

func (d DNSServer) Shutdown() error {
	for _, server := range d.Servers {
		if err := server.Shutdown(); err != nil {
			return err
		}
		d.Logger.WithFields(logrus.Fields{
			"Net":  server.Net,
			"Addr": server.Addr,
		}).Debug("successfully shut down dns server")
	}

	return nil
}

func (d DNSServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) < 1 { // allow more than one question
		dns.HandleFailed(w, req)
		return
	}

	if d.shouldBlock(req) {
		if err := d.block(w, req); err != nil {
			d.Logger.WithError(err).Error("error blocking request")
			dns.HandleFailed(w, req)
		}
		return
	}

	if err := d.forward(w, req); err != nil {
		d.Logger.WithError(err).Error("error forwarding request")
		dns.HandleFailed(w, req)
	}
}

func stripTrailing(text string, trail string) string {
	i := strings.LastIndex(text, trail)
	if i == -1 {
		return text
	}

	return text[:i]
}

func (d DNSServer) shouldBlock(req *dns.Msg) bool {
	q := req.Question[0]

	if !shouldRespondToQtype(q.Qtype) {
		return false
	}

	host := stripTrailing(q.Name, ".")

	for _, p := range d.Block.Passers {
		if p.Pass(host) {
			return false
		}
	}

	for _, b := range d.Block.Blockers {
		if b.Block(host) {
			return true
		}
	}

	return false
}

func newHdr(name string, qtype uint16, ttl uint32) dns.RR_Header {
	return dns.RR_Header{
		Name:   name,
		Rrtype: qtype,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
}

func newA(ipv4 net.IP, hdr dns.RR_Header) *dns.Msg {
	return &dns.Msg{
		Answer: []dns.RR{
			&dns.A{Hdr: hdr, A: ipv4},
		},
	}
}

func newAAAA(ipv6 net.IP, hdr dns.RR_Header) *dns.Msg {
	return &dns.Msg{
		Answer: []dns.RR{
			&dns.AAAA{Hdr: hdr, AAAA: ipv6},
		},
	}
}

func newCNAME(host string, ipv4, ipv6 net.IP, hdr dns.RR_Header) *dns.Msg {
	return &dns.Msg{
		Answer: []dns.RR{
			&dns.CNAME{Hdr: hdr, Target: host},
		},
		Extra: []dns.RR{
			&dns.A{Hdr: newHdr(host, dns.TypeA, hdr.Ttl), A: ipv4},
			&dns.AAAA{Hdr: newHdr(host, dns.TypeAAAA, hdr.Ttl), AAAA: ipv6},
		},
	}
}

func (d DNSServer) newReply(req *dns.Msg) *dns.Msg {
	q := req.Question[0]
	qType := q.Qtype
	hdr := newHdr(q.Name, qType, uint32(d.Block.TTL.Seconds()))

	var msg *dns.Msg
	switch qType {
	case dns.TypeA:
		msg = newA(d.Block.IPv4, hdr)
	case dns.TypeAAAA:
		msg = newAAAA(d.Block.IPv6, hdr)
	case dns.TypeCNAME:
		msg = newCNAME(d.Block.Host, d.Block.IPv4, d.Block.IPv6, hdr)
	default:
		panic(fmt.Sprintf("unexpected question type: %d", qType))
	}

	return msg.SetReply(req)
}

func (d DNSServer) block(w dns.ResponseWriter, req *dns.Msg) error {
	return w.WriteMsg(d.newReply(req))
}

func (d DNSServer) forward(w dns.ResponseWriter, req *dns.Msg) error {
	c := &dns.Client{Net: "udp"}
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		c.Net = "tcp"
	}

	for _, server := range d.Forward {
		resp, _, err := c.Exchange(req, server)
		if err == nil {
			// got response from forward server
			return w.WriteMsg(resp)
		}

		d.Logger.WithError(err).Error("error forwarding request")
	}

	return fmt.Errorf("all forward servers failed")
}
