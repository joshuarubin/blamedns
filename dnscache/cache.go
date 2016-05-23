package dnscache

import "github.com/miekg/dns"

type Cache interface {
	Get(*dns.Msg) *dns.Msg
	Set(*dns.Msg) int
	Prune() int
	Len() int
	Purge()
}
