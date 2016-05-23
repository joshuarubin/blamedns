package dnscache

import (
	"time"

	"github.com/Sirupsen/logrus"
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

	c.mu.RLock()
	if e, ok := c.nxDomain[q.Name]; ok {
		// it already exists, only update to lower the ttl
		if e.Expires.Before(expires) {
			c.mu.RUnlock()
			return
		}
	}

	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()

	e := &negativeEntry{
		SOA:     soa.Header().Name,
		Expires: expires,
	}
	c.nxDomain[q.Name] = e

	c.Logger.WithFields(logrus.Fields{
		"name": q.Name,
		"soa":  e.SOA,
		"ttl":  e.TTL(),
	}).Debug("stored nxdomain response in cache")
}

func (c *Memory) getNxDomain(name string) *negativeEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.nxDomain[name]; ok {
		if !e.Expired() {
			return e
		}

		delete(c.nxDomain, name)
	}

	return nil
}

func (c *Memory) buildNXReply(req *dns.Msg, e *negativeEntry) *dns.Msg {
	resp := &dns.Msg{}
	resp.SetRcode(req, dns.RcodeNameError)

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
