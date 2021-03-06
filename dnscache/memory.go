package dnscache

import (
	"context"
	"sync"
	"time"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"

	"github.com/hashicorp/golang-lru"
	"github.com/miekg/dns"
)

type key struct {
	Host string
	Type uint16
}

type lrui interface {
	Keys() []interface{}
	Get(key interface{}) (interface{}, bool)
	Remove(key interface{})
	Add(key, value interface{})
	Purge()
}

type Memory struct {
	mu     sync.Mutex
	Logger slog.Interface
	cache  lrui
	stopCh chan struct{}
}

func NewMemory(size int, logger slog.Interface) *Memory {
	if logger == nil {
		logger = text.Logger(slog.InfoLevel)
	}

	cache, err := lru.NewARC(size)
	if err != nil {
		logger.WithError(err).Panic("error creating dns memory cache")
	}

	return &Memory{
		Logger: logger,
		cache:  cache,
	}
}

func (c *Memory) Len() int {
	var n int

	for _, key := range c.cache.Keys() {
		value, ok := c.cache.Get(key)
		if !ok {
			continue
		}

		switch v := value.(type) {
		case *RRSet:
			n += v.Len()
		case *negativeEntry:
			n++
		default:
			c.Logger.WithField("value", value).Panic("invalid type stored in dns memory cache")
		}
	}

	return n
}

func (c *Memory) Prune() int {
	// there is a race between getting a value and removing it, however, it is
	// not something that can cause a panic and we prefer not to have to lock
	// the whole Prune function

	var n int
	for _, key := range c.cache.Keys() {
		value, ok := c.cache.Get(key)
		if !ok {
			continue
		}

		switch v := value.(type) {
		case *RRSet:
			n += v.Prune()
			if v.Len() == 0 {
				c.cache.Remove(key)
			}
		case *negativeEntry:
			if v.Expired() {
				n++
				c.cache.Remove(key)
			}
		default:
			c.Logger.WithField("value", value).Panic("invalid type stored in dns memory cache")
		}
	}

	return n
}

func (c *Memory) add(rr *RR) {
	k := key{
		Host: rr.Header().Name,
		Type: rr.Header().Rrtype,
	}

	// first try getting without the lock

	if value, ok := c.cache.Get(k); ok {
		if set, ok := value.(*RRSet); ok {
			set.Add(rr)
			return
		}
	}

	// wasn't there the first time, so now, get the lock and check again

	c.mu.Lock()
	defer c.mu.Unlock()

	if value, ok := c.cache.Get(k); ok {
		if set, ok := value.(*RRSet); ok {
			set.Add(rr)
			return
		}
	}

	// still wasn't there, add it

	set := &RRSet{}
	set.Add(rr)

	// this may clobber a negativeEntry, that's OK
	c.cache.Add(k, set)
}

func (c *Memory) Purge() {
	c.Logger.Info("purging dns cache")
	c.cache.Purge()
}

func (c *Memory) Set(resp *dns.Msg) int {
	if resp == nil || len(resp.Question) == 0 {
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
	fieldAnswer msgField = iota
	fieldNs
	fieldExtra
)

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

func (c *Memory) get(ctx context.Context, q dns.Question, resp *dns.Msg, field msgField) bool {
	if q.Qclass != dns.ClassINET {
		return false
	}

	if elem, ok := c.cache.Get(key{Host: q.Name, Type: q.Qtype}); ok {
		if set, ok := elem.(*RRSet); ok {
			if data := appendResponseField(resp, field, set.RR()); len(data) > 0 {
				return true
			}
		}
		return false
	}

	if q.Qtype == dns.TypeCNAME {
		return false
	}

	// there was nothing in the cache that matched the RR Type requested, see if
	// there is a CNAME that resolves to the correct type

	if elem, ok := c.cache.Get(key{Host: q.Name, Type: dns.TypeCNAME}); ok {
		if set, ok := elem.(*RRSet); ok {
			newValues := appendResponseField(resp, field, set.RR())

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
				if c.outerGet(ctx, req, resp) {
					completed = true
				}
			}

			if completed {
				return true
			}
		}
	}

	return false
}

type cacheFn struct {
	fn    func(dns.Question) *negativeEntry
	build func(context.Context, *dns.Msg, *negativeEntry, *dns.Msg) bool
}

func (c *Memory) cacheFn() []cacheFn {
	return []cacheFn{{
		fn:    c.getNoData,
		build: c.buildNegativeReply(dns.RcodeSuccess),
	}, {
		fn:    c.getNxDomain,
		build: c.buildNegativeReply(dns.RcodeNameError),
	}}
}

func (c *Memory) Get(ctx context.Context, req *dns.Msg) *dns.Msg {
	if ctx == nil {
		ctx = context.Background()
	}

	resp := &dns.Msg{}
	if c.outerGet(ctx, req, resp) {
		return resp
	}
	return nil
}

func (c *Memory) outerGet(ctx context.Context, req *dns.Msg, resp *dns.Msg) bool {
	// check to see if ctx was canceled before spinning up any requests since
	// c.get may call outerGet again
	select {
	case <-ctx.Done():
		return false
	default:
	}

	q := req.Question[0]

	getCh := make(chan bool, 1)

	go func() {
		if c.get(ctx, q, resp, fieldAnswer) {
			resp.SetReply(req)
			c.processAdditionalSection(ctx, resp)
			// TODO(jrubin) ensure no duplicates of resp.Answer are in resp.Extra
			getCh <- true
			return
		}

		// check for NXDOMAIN and NODATA

		for _, fn := range c.cacheFn() {
			if e := fn.fn(q); e != nil {
				if fn.build(ctx, req, e, resp) {
					getCh <- true
					return
				}
			}
		}

		getCh <- false
	}()

	select {
	case <-ctx.Done():
		return false
	case got := <-getCh:
		return got
	}
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

func (c *Memory) processAdditionalSection(ctx context.Context, resp *dns.Msg) {
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
		c.get(ctx, q, resp, fieldExtra)
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
