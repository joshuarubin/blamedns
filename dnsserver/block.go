package dnsserver

import (
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
	TTL        time.Duration
	Blocker    Blocker
	Passer     Passer
	Logger     *logrus.Logger
}

func unfqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}

func newHdr(name string, qtype uint16, ttl uint32) dns.RR_Header {
	return dns.RR_Header{
		Name:   name,
		Rrtype: qtype,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
}

func (b Block) Should(req *dns.Msg) bool {
	q := req.Question[0]

	switch q.Qtype {
	case dns.TypeA, dns.TypeAAAA:
		// noop, only respond to these query types
	default:
		return false
	}

	host := strings.ToLower(unfqdn(q.Name))

	if b.Passer.Pass(host) {
		return false
	}

	if b.Blocker.Block(host) {
		return true
	}

	return false
}

func (b Block) NewReply(req *dns.Msg) *dns.Msg {
	q := req.Question[0]
	hdr := newHdr(q.Name, q.Qtype, uint32(b.TTL.Seconds()))

	var rr dns.RR
	switch q.Qtype {
	case dns.TypeA:
		rr = &dns.A{Hdr: hdr, A: b.IPv4}
	case dns.TypeAAAA:
		rr = &dns.AAAA{Hdr: hdr, AAAA: b.IPv6}
	default:
		b.Logger.WithFields(logrus.Fields{
			"type": dns.TypeToString[q.Qtype],
		}).Panic("unexpected question type")
	}

	msg := &dns.Msg{
		Answer: []dns.RR{rr},
	}

	return msg.SetReply(req)
}
