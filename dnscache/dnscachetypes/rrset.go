package dnscachetypes

import "github.com/miekg/dns"

type RRSet []*RR

func (rs *RRSet) Add(rr dns.RR) *RR {
	r := NewRR(rr)

	var added bool
	ret := make(RRSet, 0, len(*rs)+1)
	for _, t := range *rs {
		// might as well prune since we are here
		if t.Expired() {
			continue
		}

		if !t.Equal(r) {
			ret = append(ret, t)
			continue
		}

		added = true

		// found equal rr already in cache
		// keep only the one with the lower ttl

		n := r
		if t.Expires.Before(r.Expires) {
			n = t
		}

		ret = append(ret, n)
	}

	if !added {
		ret = append(ret, r)
	}

	*rs = ret

	return r
}

func (rs RRSet) RR() []dns.RR {
	var ret []dns.RR
	for _, rr := range rs {
		if !rr.Expired() {
			r := dns.Copy(rr.RR)
			r.Header().Ttl = rr.TTL().Seconds()
			ret = append(ret, r)
		}
	}
	return ret
}

func (rs *RRSet) Prune() int {
	var n int
	ret := make(RRSet, 0, len(*rs))
	for _, rr := range *rs {
		if rr.Expired() {
			n++
			continue
		}

		ret = append(ret, rr)
	}
	*rs = ret
	return n
}

func (rs RRSet) Len() int {
	return len(rs)
}
