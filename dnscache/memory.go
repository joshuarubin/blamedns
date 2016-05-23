package dnscache

import (
	"container/list"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

type key struct {
	Host string
	Type uint16
}

type entry struct {
	Key key
	Set *RRSet
}

type EvictNotifier interface {
	NotifyEvicted(typ uint16, host string, rr []dns.RR)
}

type EvictNotifierFunc func(typ uint16, host string, rr []dns.RR)

func (f EvictNotifierFunc) NotifyEvicted(typ uint16, host string, rr []dns.RR) {
	f(typ, host, rr)
}

type Memory struct {
	mu        sync.RWMutex
	size      int
	evictList *list.List
	data      map[key]*list.Element
	Logger    *logrus.Logger
	onEvict   EvictNotifier
}

func NewMemory(size int, logger *logrus.Logger, onEvict EvictNotifier) *Memory {
	if size <= 0 {
		panic("dns memory cache must have a positive size")
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &Memory{
		size:      size,
		evictList: list.New(),
		data:      map[key]*list.Element{},
		Logger:    logger,
		onEvict:   onEvict,
	}
}

func (c *Memory) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var n int
	for _, elem := range c.data {
		n += elem.Value.(*entry).Set.Len()
	}
	return n
}

func (c *Memory) Prune() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var n int
	for _, elem := range c.data {
		set := elem.Value.(*entry).Set
		n += set.Prune()
		if set.Len() == 0 {
			c.removeElement(elem)
		}
	}
	return n
}

func (c *Memory) add(rr *RR) *RR {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := key{
		Host: rr.Header().Name,
		Type: rr.Header().Rrtype,
	}

	elem, ok := c.data[k]
	var set *RRSet
	if ok {
		c.evictList.MoveToFront(elem)
		set = elem.Value.(*entry).Set
	} else {
		set = &RRSet{}
		elem = c.evictList.PushFront(&entry{Key: k, Set: set})
		c.data[k] = elem
	}

	if c.evictList.Len() > c.size {
		c.removeOldest()
	}

	return set.Add(rr)
}

func (c *Memory) removeOldest() {
	if elem := c.evictList.Back(); elem != nil {
		c.removeElement(elem)
	}
}

func (c *Memory) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)

	ent := elem.Value.(*entry)
	delete(c.data, ent.Key)
	if c.onEvict != nil {
		c.onEvict.NotifyEvicted(ent.Key.Type, ent.Key.Host, ent.Set.RR())
	}
}

func (c *Memory) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, elem := range c.data {
		if c.onEvict != nil {
			c.onEvict.NotifyEvicted(k.Type, k.Host, elem.Value.(*entry).Set.RR())
		}
		delete(c.data, k)
	}
	c.evictList.Init()
}

func (c *Memory) Set(resp *dns.Msg) int {
	if c == nil {
		return 0
	}

	if len(resp.Question) != 1 {
		return 0
	}

	q := resp.Question[0]
	if q.Qclass != dns.ClassINET {
		return 0
	}

	// TODO(jrubin) implement negative response caching according to:
	// https://tools.ietf.org/html/rfc2308#section-5

	if resp.Rcode == dns.RcodeNameError {
		// NXDOMAIN
	} else if resp.Rcode != dns.RcodeSuccess {
		return 0
	} else if len(resp.Answer) == 0 {
		// NODATA
	}

	var n int
	for _, s := range [][]dns.RR{resp.Answer, resp.Ns, resp.Extra} {
		for _, rr := range s {
			if rr.Header().Class != dns.ClassINET {
				continue
			}

			e := c.add(NewRR(rr))
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

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key{Host: q.Name, Type: q.Qtype}]; ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*entry).Set.RR()
	}

	return nil
}

func (c *Memory) Get(req *dns.Msg) *dns.Msg {
	if c == nil || len(req.Question) != 1 {
		return nil
	}

	q := req.Question[0]

	if ans := c.get(q); ans != nil {
		c.Logger.WithFields(logFields(q, "hit")).Debug("cache lookup")
		return c.buildReply(req, ans)
	}

	c.Logger.WithFields(logFields(q, "miss")).Debug("cache lookup")
	return nil
}

func (c *Memory) addDNSSEC(req, resp *dns.Msg) {
	// add rrsig records (regardless of DNSSEC OK flag)
	q := req.Copy().Question[0]
	q.Qtype = dns.TypeRRSIG

	var lf logrus.Fields
	if ans := c.get(q); ans != nil {
		lf = logFields(q, "hit")
		resp.Answer = append(resp.Answer, ans...)
	} else {
		lf = logFields(q, "miss")
	}

	c.Logger.WithFields(lf).Debug("cache lookup (rrsig)")
}

func (c *Memory) buildReply(req *dns.Msg, ans []dns.RR) *dns.Msg {
	resp := &dns.Msg{}
	resp.SetReply(req)
	resp.Answer = ans

	c.addDNSSEC(req, resp)
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
