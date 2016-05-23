package dnscache

import "github.com/miekg/dns"

type Cache interface {
	Get(dns.Question) []dns.RR
	Set(*dns.Msg) int
	Prune() int
	Len() int
	FlushAll()
}
