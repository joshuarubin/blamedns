package dnsserver

import (
	"fmt"
	"net"
	"strings"
	"time"

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
	Blockers   []Blocker
	Passers    []Passer
}

func unfqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}

func (b Block) Should(req *dns.Msg) bool {
	q := req.Question[0]

	switch q.Qtype {
	case dns.TypeA, dns.TypeAAAA:
		// noop, only respond to these query types
	default:
		return false
	}

	host := unfqdn(q.Name)
	host = strings.ToLower(host)

	for _, p := range b.Passers {
		if p.Pass(host) {
			return false
		}
	}

	for _, b := range b.Blockers {
		if b.Block(host) {
			return true
		}
	}

	return false
}

func (b Block) NewReply(req *dns.Msg) *dns.Msg {
	q := req.Question[0]
	qType := q.Qtype
	hdr := newHdr(q.Name, qType, uint32(b.TTL.Seconds()))

	var msg *dns.Msg
	switch qType {
	case dns.TypeA:
		msg = newA(b.IPv4, hdr)
	case dns.TypeAAAA:
		msg = newAAAA(b.IPv6, hdr)
	default:
		panic(fmt.Sprintf("unexpected question type: %d", qType))
	}

	return msg.SetReply(req)
}
