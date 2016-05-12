package dnsserver

import (
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

type Block struct {
	IP  net.IP
	TTL time.Duration
}

type DNSServer struct {
	Forward []string
	Servers []*dns.Server
	Block   Block
}

const defaultDNSPort = 53

// TODO(jrubin) does CNAME need to be handled?
// TODO(jrubin) does AAAA need to be handled?
// TODO(jrubin) handle other question types?
var respondToQtypes = []uint16{dns.TypeA}

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

	log.Infof("using forwarders: %s", strings.Join(d.Forward, ", "))

	errCh := make(chan error, len(d.Servers))

	for _, server := range d.Servers {
		server.Handler = d
		go func(server *dns.Server) {
			log.WithFields(log.Fields{
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
		log.WithFields(log.Fields{
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

	if shouldBlock(req) {
		if err := d.block(w, req); err != nil {
			log.WithError(err).Error("error blocking request")
			dns.HandleFailed(w, req)
		}
		return
	}

	if err := d.forward(w, req); err != nil {
		log.WithError(err).Error("error forwarding request")
		dns.HandleFailed(w, req)
	}
}

func shouldBlock(req *dns.Msg) bool {
	if !shouldRespondToQtype(req.Question[0].Qtype) {
		return false
	}

	// TODO(jrubin) check if req.Question[0].Name should be blocked
	return req.Question[0].Name == "blamedns.com."
}

func (d DNSServer) block(w dns.ResponseWriter, req *dns.Msg) error {
	resp := &dns.Msg{}
	resp.SetReply(req)

	resp.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   req.Question[0].Name,
				Rrtype: req.Question[0].Qtype,
				Class:  dns.ClassINET,
				Ttl:    uint32(d.Block.TTL.Seconds()),
			},
			A: d.Block.IP,
		},
	}

	return w.WriteMsg(resp)
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

		log.WithError(err).Error("error forwarding request")
	}

	return fmt.Errorf("all forward servers failed")
}
