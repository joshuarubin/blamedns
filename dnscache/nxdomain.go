package dnscache

import (
	"time"

	"github.com/miekg/dns"
)

type negativeEntry struct {
	SOA     string
	Expires time.Time
}

func (e negativeEntry) TTL() TTL {
	return TTL(e.Expires.Sub(time.Now().UTC()))
}

func (e negativeEntry) Expired() bool {
	return e.Expires.Before(time.Now().UTC())
}

func getSOA(resp *dns.Msg) *dns.SOA {
	for _, n := range resp.Ns {
		if soa, ok := n.(*dns.SOA); ok {
			return soa
		}
	}
	return nil
}

func (c *Memory) setNxDomain(resp *dns.Msg) {
	if resp.Rcode != dns.RcodeNameError {
		// wasn't an NXDOMAIN response
		return
	}

	soa := getSOA(resp)
	if soa == nil {
		return
	}

	ttl := soa.Header().Ttl
	if soa.Minttl < ttl {
		ttl = soa.Minttl
	}

	expires := time.Now().UTC().Add(time.Duration(ttl) * time.Second)

	q := resp.Question[0]

	if value, ok := c.cache.Get(q.Name); ok {
		if ne, ok := value.(*negativeEntry); ok {
			// it already exists, only update to lower the ttl
			if ne.Expires.Before(expires) {
				return
			}
		}
	}

	e := &negativeEntry{
		SOA:     soa.Header().Name,
		Expires: expires,
	}

	c.cache.Add(q.Name, e)
}

func (c *Memory) getNxDomain(q dns.Question) *negativeEntry {
	if value, ok := c.cache.Get(q.Name); ok {
		if ne, ok := value.(*negativeEntry); ok {
			if !ne.Expired() {
				return ne
			}

			c.cache.Remove(q.Name)
		}
	}

	return nil
}
