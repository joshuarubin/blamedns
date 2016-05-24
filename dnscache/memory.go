package dnscache

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"jrubin.io/blamedns/lru"
)

type key struct {
	Host string
	Type uint16
}

type Memory struct {
	Logger   *logrus.Logger
	data     *lru.LRU
	nxDomain *lru.LRU
	noData   *lru.LRU
	stopCh   chan struct{}
}

func NewMemory(size int, logger *logrus.Logger, onEvict lru.Elementer) *Memory {
	if logger == nil {
		logger = logrus.New()
	}

	// TODO(jrubin) should each lru be sized differently?
	return &Memory{
		Logger:   logger,
		data:     lru.New(size, onEvict),
		nxDomain: lru.New(size, onEvict),
		noData:   lru.New(size, onEvict),
	}
}

func (c *Memory) Len() int {
	var n int

	for _, l := range []*lru.LRU{c.data, c.nxDomain, c.noData} {
		l.Each(lru.ElementerFunc(func(k, value interface{}) {
			n += value.(*RRSet).Len()
		}))
	}

	return n
}

func (c *Memory) Prune() int {
	var n int
	c.data.Each(lru.ElementerFunc(func(k, value interface{}) {
		set := value.(*RRSet)
		n += set.Prune()
		if set.Len() == 0 {
			c.data.RemoveNoLock(k)
		}
	}))

	for _, l := range []*lru.LRU{c.nxDomain, c.noData} {
		l.Each(lru.ElementerFunc(func(k, value interface{}) {
			if value.(*negativeEntry).Expired() {
				n++
				l.RemoveNoLock(k)
			}
		}))
	}

	return n
}

func (c *Memory) add(rr *RR) {
	k := key{
		Host: rr.Header().Name,
		Type: rr.Header().Rrtype,
	}

	value, ok := c.data.Get(k)
	var set *RRSet
	if ok {
		set = value.(*RRSet)
	} else {
		set = &RRSet{}
		if !c.data.AddIfNotContains(k, set) {
			c.add(rr)
			return
		}
	}

	c.data.Lock()
	defer c.data.Unlock()
	set.Add(rr)
}

func (c *Memory) Purge() {
	for _, lru := range []*lru.LRU{c.data, c.nxDomain, c.noData} {
		lru.Purge()
	}
}

func (c *Memory) Set(resp *dns.Msg) int {
	q := resp.Question[0]
	if q.Qclass != dns.ClassINET {
		return 0
	}

	// cache negative responses according to:
	// https://tools.ietf.org/html/rfc2308#section-5

	if resp.Rcode == dns.RcodeNameError {
		// NXDOMAIN cache for all Name regardless of Qtype
		c.setNxDomain(resp)
	} else if resp.Rcode != dns.RcodeSuccess {
		return 0
	} else if len(resp.Answer) == 0 {
		// NODATA cache for only Qtype/Name conbination
		c.setNoData(resp)
	}

	var n int
	for _, s := range [][]dns.RR{resp.Answer, resp.Ns, resp.Extra} {
		for _, rr := range s {
			if rr.Header().Class != dns.ClassINET {
				continue
			}

			e := NewRR(rr)
			c.add(e)
			n++

			c.Logger.WithFields(logrus.Fields{
				"type": dns.TypeToString[e.Header().Rrtype],
				"name": e.Header().Name,
				"ttl":  e.TTL().Seconds(),
			}).Debug("stored in cache")
		}
	}

	return n
}

func logFields(q dns.Question, cache string) logrus.Fields {
	return logrus.Fields{
		"class": dns.ClassToString[q.Qclass],
		"type":  dns.TypeToString[q.Qtype],
		"name":  q.Name,
		"cache": cache,
	}
}

func (c *Memory) get(q dns.Question) []dns.RR {
	if q.Qclass != dns.ClassINET {
		return nil
	}

	if elem, ok := c.data.Get(key{Host: q.Name, Type: q.Qtype}); ok {
		return elem.(*RRSet).RR()
	}

	return nil
}

func (c *Memory) Get(req *dns.Msg) *dns.Msg {
	q := req.Question[0]

	if ans := c.get(q); ans != nil {
		c.Logger.WithFields(logFields(q, "hit")).Debug("cache lookup")
		return c.buildReply(req, ans)
	}

	if e := c.getNoData(q); e != nil {
		if r := c.buildNoDataReply(req, e); r != nil {
			lf := logFields(q, "hit")
			lf["ttl"] = e.TTL().Seconds()
			c.Logger.WithFields(lf).Debug("cache lookup (nodata)")
			return r
		}
	}

	if e := c.getNxDomain(q); e != nil {
		if r := c.buildNXReply(req, e); r != nil {
			lf := logFields(q, "hit")
			lf["ttl"] = e.TTL().Seconds()
			c.Logger.WithFields(lf).Debug("cache lookup (nxdomain)")
			return r
		}
	}

	c.Logger.WithFields(logFields(q, "miss")).Debug("cache lookup")
	return nil
}

func (c *Memory) addDNSSEC(resp *dns.Msg) {
	// add rrsig records (regardless of DNSSEC OK flag)
	for i, set := range [][]dns.RR{resp.Answer, resp.Ns, resp.Extra} {
		for _, rr := range set {
			if rr.Header().Rrtype == dns.TypeRRSIG {
				continue
			}

			q := dns.Question{
				Name:   rr.Header().Name,
				Qtype:  dns.TypeRRSIG,
				Qclass: dns.ClassINET,
			}

			var lf logrus.Fields
			if ans := c.get(q); ans != nil {
				lf = logFields(q, "hit")

				switch i {
				case 0: // Answer
					resp.Answer = append(resp.Answer, ans...)
				case 1: // Ns
					resp.Ns = append(resp.Ns, ans...)
				case 2: // Extra
					resp.Extra = append(resp.Extra, ans...)
				}
			} else {
				lf = logFields(q, "miss")
			}
			c.Logger.WithFields(lf).Debug("cache lookup (rrsig)")
		}
	}
}

func (c *Memory) buildReply(req *dns.Msg, ans []dns.RR) *dns.Msg {
	resp := &dns.Msg{}
	resp.SetReply(req)
	resp.Answer = ans

	c.addDNSSEC(resp)
	c.processAdditionalSection(resp)

	// TODO(jrubin) ensure no duplicates of resp.Answer are in resp.Extra

	return resp
}

var additionalQTypes = []uint16{dns.TypeA, dns.TypeAAAA}

func additionalQ(qs []dns.Question, name string) []dns.Question {
	for _, t := range additionalQTypes {
		qs = append(qs, dns.Question{
			Name:   name,
			Qtype:  t,
			Qclass: dns.ClassINET,
		})
	}
	return qs
}

func (c *Memory) processAdditionalSection(resp *dns.Msg) {
	// CNAMEs do not cause "additional section processing"
	// https://tools.ietf.org/html/rfc2181#section-10.3
	//
	// Not Implemented:
	// MB (experimental)
	// MD (obsolete)
	// MF (obsolete)
	// Sig?

	var qs []dns.Question

	for _, rr := range resp.Answer {
		switch t := rr.(type) {
		case *dns.NS:
			qs = additionalQ(qs, t.Ns)
		case *dns.MX:
			qs = additionalQ(qs, t.Mx)
		case *dns.SRV:
			qs = additionalQ(qs, t.Target)
		}
	}

	for _, q := range qs {
		var lf logrus.Fields
		if extra := c.get(q); extra != nil {
			lf = logFields(q, "hit")
			resp.Extra = append(resp.Extra, extra...)
		} else {
			lf = logFields(q, "miss")
		}

		c.Logger.WithFields(lf).Debug("cache lookup (additional)")
	}
}

func (c *Memory) Start(i time.Duration) {
	c.Logger.WithField("interval", i).Info("started cache prune background process")
	ticker := time.NewTicker(i)
	c.stopCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Logger.WithField("num", c.Prune()).Debug("pruned items from cache")
			case <-c.stopCh:
				ticker.Stop()
				c.Logger.Debug("stopped cache pruner")
				return
			}
		}
	}()
}

func (c *Memory) Stop() {
	if c.stopCh != nil {
		c.stopCh <- struct{}{}
		close(c.stopCh)
		c.stopCh = nil
	}
}
