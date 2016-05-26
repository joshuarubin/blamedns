package dnscache

import (
	"fmt"
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
	if resp == nil {
		return 0
	}

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
	} else if respNODATA(resp) {
		// NODATA cache for only Qtype/Name conbination
		c.setNoData(resp)
	}

	var n int
	for _, s := range [][]dns.RR{resp.Answer, resp.Ns, resp.Extra} {
		for _, rr := range s {
			if rr.Header().Class == dns.ClassINET {
				c.add(NewRR(rr))
				n++
			}
		}
	}
	return n
}

type msgField int

const (
	fieldAnswer = iota
	fieldNs
	fieldExtra
)

func responseField(resp *dns.Msg, field msgField) []dns.RR {
	switch field {
	case fieldAnswer:
		return resp.Answer
	case fieldNs:
		return resp.Ns
	case fieldExtra:
		return resp.Extra
	}
	panic(fmt.Errorf("invalid field: %d", field))
}

func appendResponseField(resp *dns.Msg, field msgField, value []dns.RR) []dns.RR {
	var newValues []dns.RR
	switch field {
	case fieldAnswer:
		resp.Answer, newValues = appendRRSetIfNotExist(resp.Answer, value)
	case fieldNs:
		resp.Ns, newValues = appendRRSetIfNotExist(resp.Ns, value)
	case fieldExtra:
		resp.Extra, newValues = appendRRSetIfNotExist(resp.Extra, value)
	}
	return newValues
}

func (c *Memory) get(q dns.Question, resp *dns.Msg, field msgField) bool {
	if q.Qclass != dns.ClassINET {
		return false
	}

	if elem, ok := c.data.Get(key{Host: q.Name, Type: q.Qtype}); ok {
		if data := appendResponseField(resp, field, elem.(*RRSet).RR()); len(data) > 0 {
			return true
		}
		return false
	}

	if q.Qtype == dns.TypeCNAME {
		return false
	}

	// there was nothing in the cache that matched the RR Type requested, see if
	// there is a CNAME that resolves to the correct type

	if elem, ok := c.data.Get(key{Host: q.Name, Type: dns.TypeCNAME}); ok {
		newValues := appendResponseField(resp, field, elem.(*RRSet).RR())

		var completed bool
		for _, value := range newValues {
			cname := value.(*dns.CNAME)

			req := &dns.Msg{
				Question: []dns.Question{{
					Name:   cname.Target,
					Qtype:  q.Qtype,
					Qclass: q.Qclass,
				}},
			}

			// we need to use the outside Get() so that NXDOMAIN and NODATA
			// values are correct as well
			if c.outerGet(req, resp) {
				completed = true
			}
		}

		if completed {
			return true
		}
	}

	return false
}

type cacheFn struct {
	fn    func(dns.Question) *negativeEntry
	build func(*dns.Msg, *negativeEntry, *dns.Msg) bool
	name  string
}

func (c *Memory) cacheFn() []cacheFn {
	return []cacheFn{{
		fn:    c.getNoData,
		build: c.buildNegativeReply(dns.RcodeSuccess),
		name:  "nodata",
	}, {
		fn:    c.getNxDomain,
		build: c.buildNegativeReply(dns.RcodeNameError),
		name:  "nxdomain",
	}}
}

func (c *Memory) Get(req *dns.Msg) *dns.Msg {
	resp := &dns.Msg{}
	if c.outerGet(req, resp) {
		return resp
	}
	return nil
}

func (c *Memory) outerGet(req *dns.Msg, resp *dns.Msg) bool {
	q := req.Question[0]

	if c.get(q, resp, fieldAnswer) {
		resp.SetReply(req)
		c.processAdditionalSection(resp)
		// TODO(jrubin) ensure no duplicates of resp.Answer are in resp.Extra
		return true
	}

	// check for NXDOMAIN and NODATA

	for _, fn := range c.cacheFn() {
		if e := fn.fn(q); e != nil {
			if fn.build(req, e, resp) {
				return true
			}
		}
	}

	return false
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
		c.get(q, resp, fieldExtra)
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
