package dnscache

import "github.com/miekg/dns"

type RRSet []*RR

func (rs *RRSet) Add(r *RR) {
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
	(*rs)[len((*rs))-1] = nil
	*rs = (*rs)[:len((*rs))-1]
}

func (rs *RRSet) RR() []dns.RR {
	var deleted int
	ret := make([]dns.RR, 0, len(*rs))
	for i, rr := range *rs {
		j := i - deleted
		if _, ok := rs.prune(j); ok {
			deleted++
			continue
		}
		ret = append(ret, rr.RR())
	}

	if len(ret) == 0 {
		return nil
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
	var ret int
	for _, rr := range rs {
		if !rr.Expired() {
			ret++
		}
	}
	return ret
}

func appendRRSetIfNotExist(data, set []dns.RR) ([]dns.RR, []dns.RR) {
	var newlyAppended []dns.RR

	for _, rr := range set {
		var appended bool
		if data, appended = appendRRIfNotExist(data, rr); appended {
			newlyAppended = append(newlyAppended, rr)
		}
	}

	return data, newlyAppended
}

func appendRRIfNotExist(data []dns.RR, val dns.RR) ([]dns.RR, bool) {
	rr := NewRR(val)

	var found bool
	for _, exist := range data {
		if rr.Equal(NewRR(exist)) {
			found = true
			break
		}
	}

	if !found {
		data = append(data, val)
	}

	return data, !found
}
