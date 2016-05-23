package dnscachetypes

import "github.com/miekg/dns"

type RRSet []*RR

func (rs *RRSet) Add(rr dns.RR) *RR {
	r := NewRR(rr)

	var deleted int
	var added bool
	for i := range *rs {
		j := i - deleted

		// might as well prune since we are here
		t, pruned := rs.prune(j)

		if pruned {
			deleted++
			continue
		}

		if !t.Equal(r) {
			continue
		}

		added = true

		// found equal rr already in cache
		// keep only the one with the lower ttl

		if r.Expires.Before(t.Expires) {
			rs.set(j, r)
		}
	}

	if !added {
		rs.append(r)
	}

	return r
}

func (rs *RRSet) get(i int) *RR {
	return (*rs)[i]
}

func (rs *RRSet) set(i int, r *RR) {
	(*rs)[i] = r
}

func (rs *RRSet) append(rr *RR) {
	*rs = append(*rs, rr)
}

func (rs *RRSet) delete(i int) {
	copy((*rs)[i:], (*rs)[i+1:])
	(*rs)[len((*rs))-1] = nil // or the zero value of T
	*rs = (*rs)[:len((*rs))-1]
}

func (rs RRSet) RR() []dns.RR {
	ret := make([]dns.RR, 0, rs.Len())
	for _, rr := range rs {
		if !rr.Expired() {
			ret = append(ret, rr.RR())
		}
	}
	return ret
}

func (rs *RRSet) prune(i int) (*RR, bool) {
	rr := rs.get(i)
	if rr.Expired() {
		rs.delete(i)
		return rr, true
	}
	return rr, false
}

func (rs *RRSet) Prune() int {
	var deleted int
	for i := range *rs {
		j := i - deleted
		if _, ok := rs.prune(j); ok {
			deleted++
		}
	}
	return deleted
}

func (rs RRSet) Len() int {
	return len(rs)
}
