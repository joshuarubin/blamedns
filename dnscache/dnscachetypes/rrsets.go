package dnscachetypes

import "github.com/miekg/dns"

type RRSets map[uint16]*RRSet

func (rss RRSets) Add(rr dns.RR) *RR {
	t := rr.Header().Rrtype
	rs, ok := rss[t]
	if !ok {
		rs = &RRSet{}
		rss[t] = rs
	}
	return rs.Add(rr)
}

func (rss RRSets) Get(q dns.Question) []dns.RR {
	rs, ok := rss[q.Qtype]
	if !ok {
		return nil
	}
	return rs.RR()
}

func (rss RRSets) Prune() int {
	var n int
	for _, set := range rss {
		n += set.Prune()
	}
	return n
}

func (rss RRSets) Len() int {
	var n int
	for _, set := range rss {
		n += set.Len()
	}
	return n
}
