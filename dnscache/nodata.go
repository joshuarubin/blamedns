package dnscache

import (
	"context"
	"time"

	"github.com/miekg/dns"
)

func respNODATA(resp *dns.Msg) bool {
	// 3 types of NODATA:
	// https://tools.ietf.org/html/rfc2308
	// RCODE always NOERROR
	//
	// 1. No, or only CNAME Answers. Authority has SOA ans NS.
	// 2. No, or only CNAME Answers, Authority just SOA.
	// 3. No, or only CNAME Answers. No Authority. REFERRAL RESPONSE (not
	//    handled here)
	//
	// here is a scenario where this happens record exists for CNAME
	// sub.example.com -> srv.example.com srv.example.com has an A record, but
	// not AAAA a query for AAAA sub.example.com should return a CNAME answer
	// and authority data

	if resp.Rcode != dns.RcodeSuccess {
		// a failure should not be cached
		return false
	}

	if soa := getSOA(resp); soa == nil {
		// no soa, nothing we can cache on
		return false
	}

	if len(resp.Answer) == 0 {
		// success. soa. no answers. nodata.
		return true
	}

	if resp.Question[0].Qtype == dns.TypeCNAME {
		// there were answers, so if it was a CNAME request it probably should
		// have already been cached normally
		return false
	}

	// check the type of the last answer
	if resp.Answer[len(resp.Answer)-1].Header().Rrtype == dns.TypeCNAME {
		// there were answers, but they were all CNAMEs (and the question
		// wasn't). nodata.
		return true
	}

	return false
}

func (c *Memory) setNoData(resp *dns.Msg) {
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
	k := key{
		Host: q.Name,
		Type: q.Qtype,
	}

	if value, ok := c.cache.Get(k); ok {
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

	c.cache.Add(k, e)
}

func (c *Memory) getNoData(q dns.Question) *negativeEntry {
	k := key{
		Host: q.Name,
		Type: q.Qtype,
	}

	if value, ok := c.cache.Get(k); ok {
		if ne, ok := value.(*negativeEntry); ok {
			if !ne.Expired() {
				return ne
			}

			c.cache.Remove(k)
		}
	}

	return nil
}

func (c *Memory) buildNegativeReply(rcode int) func(context.Context, *dns.Msg, *negativeEntry, *dns.Msg) bool {
	return func(ctx context.Context, req *dns.Msg, e *negativeEntry, resp *dns.Msg) bool {
		q := dns.Question{
			Name:   e.SOA,
			Qtype:  dns.TypeSOA,
			Qclass: dns.ClassINET,
		}

		if c.get(ctx, q, resp, fieldNs) {
			resp.SetRcode(req, rcode)
			return true
		}

		return false
	}
}
