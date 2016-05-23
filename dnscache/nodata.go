package dnscache

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func (c *Memory) setNoData(resp *dns.Msg) {
	if resp.Rcode != dns.RcodeSuccess || len(resp.Answer) > 0 {
		// wasn't an NODATA response
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
	k := key{
		Host: q.Name,
		Type: q.Qtype,
	}

	if e, ok := c.noData[k]; ok {
		// it already exists, only update to lower the ttl
		if e.Expires.Before(expires) {
			return
		}
	}

	e := &negativeEntry{
		SOA:     soa.Header().Name,
		Expires: expires,
	}
	c.noData[k] = e

	c.Logger.WithFields(logrus.Fields{
		"name": q.Name,
		"soa":  e.SOA,
		"ttl":  e.TTL().Seconds(),
	}).Debug("stored nodata response in cache")
}

func (c *Memory) getNoData(q dns.Question) *negativeEntry {
	k := key{
		Host: q.Name,
		Type: q.Qtype,
	}

	if e, ok := c.noData[k]; ok {
		if !e.Expired() {
			return e
		}

		delete(c.noData, k)
	}

	return nil
}

func (c *Memory) buildNoDataReply(req *dns.Msg, e *negativeEntry) *dns.Msg {
	resp := &dns.Msg{}
	resp.SetReply(req)

	q := dns.Question{
		Name:   e.SOA,
		Qtype:  dns.TypeSOA,
		Qclass: dns.ClassINET,
	}

	if resp.Ns = c.get(q); resp.Ns != nil {
		c.addDNSSEC(resp)
		return resp
	}

	return nil
}
