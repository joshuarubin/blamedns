package dnsserver

import (
	"net"

	"github.com/miekg/dns"
)

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
