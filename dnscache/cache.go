package dnscache

import "github.com/miekg/dns"

type Cache interface {
	Get(dns.Question) []dns.RR
	Set(*dns.Msg) int
	Prune() int
	Len() int
	FlushAll()
}

func unfqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
