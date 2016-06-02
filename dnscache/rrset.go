package dnscache

import (
	"sync"

	"github.com/miekg/dns"
)

type RRSet struct {
	sync.Mutex
	data []*RR
}

func (rs *RRSet) Add(r *RR) {
	rs.Lock()
	defer rs.Unlock()

	var deleted int
	var added bool
	for i := range rs.data {
		j := i - deleted

		// might as well prune since we are here
		t, pruned := rs.pruneNoLock(j)

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
			rs.setNoLock(i, r)
		}
	}

	if !added {
		rs.appendNoLock(r)
	}
}

func (rs *RRSet) getNoLock(i int) *RR {
	return rs.data[i]
}

func (rs *RRSet) setNoLock(i int, r *RR) {
	rs.data[i] = r
}

func (rs *RRSet) appendNoLock(rr *RR) {
	rs.data = append(rs.data, rr)
}

func (rs *RRSet) deleteNoLock(i int) {
	copy(rs.data[i:], rs.data[i+1:])
	rs.data[len(rs.data)-1] = nil
	rs.data = rs.data[:len(rs.data)-1]
}

func (rs *RRSet) RR() []dns.RR {
	rs.Lock()
	defer rs.Unlock()

	var deleted int
	ret := make([]dns.RR, 0, len(rs.data))
	for i := range rs.data {
		j := i - deleted
		if _, ok := rs.pruneNoLock(j); ok {
			deleted++
			continue
		}
		ret = append(ret, rs.getNoLock(j).RR())
	}

	if len(ret) == 0 {
		return nil
	}

	return ret
}

func (rs *RRSet) pruneNoLock(i int) (*RR, bool) {
	rr := rs.getNoLock(i)
	if rr.Expired() {
		rs.deleteNoLock(i)
		return rr, true
	}
	return rr, false
}

func (rs *RRSet) Prune() int {
	rs.Lock()
	defer rs.Unlock()

	var deleted int
	for i := range rs.data {
		j := i - deleted
		if _, ok := rs.pruneNoLock(j); ok {
			deleted++
		}
	}
	return deleted
}

func (rs *RRSet) Len() int {
	rs.Lock()
	defer rs.Unlock()

	var ret int
	for _, rr := range rs.data {
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
