package dnscache

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

type key struct {
	Host string
	Type uint16
}

type Memory struct {
	mu     sync.RWMutex
	data   map[key]*RRSet
	Logger *logrus.Logger
}

func NewMemory(logger *logrus.Logger) *Memory {
	return &Memory{
		data:   map[key]*RRSet{},
		Logger: logger,
	}
}

var _ Cache = &Memory{}

func (c *Memory) Len() int {
	var n int
	for _, sets := range c.data {
		n += sets.Len()
	}
	return n
}

func (c *Memory) Prune() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var n int
	for _, sets := range c.data {
		n += sets.Prune()
	}
	return n
}

func (c *Memory) add(rr dns.RR) *RR {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := key{
		Host: rr.Header().Name,
		Type: rr.Header().Rrtype,
	}

	rrs, ok := c.data[k]
	if !ok || rrs == nil {
		rrs = &RRSet{}
		c.data[k] = rrs
	}
	return rrs.Add(rr)
}

func (c *Memory) FlushAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = map[key]*RRSet{}
}

func (c *Memory) Set(res *dns.Msg) int {
	if len(res.Question) != 1 {
		return 0
	}

	q := res.Question[0]
	if q.Qclass != dns.ClassINET {
		return 0
	}

	var n int
	for _, s := range [][]dns.RR{res.Answer, res.Ns, res.Extra} {
		for _, rr := range s {
			if rr.Header().Class != dns.ClassINET {
				continue
			}

			e := c.add(rr)
			n++

			if c.Logger != nil {
				c.Logger.WithFields(logrus.Fields{
					"type": dns.TypeToString[e.Header().Rrtype],
					"name": unfqdn(e.Header().Name),
					"ttl":  e.TTL().Seconds(),
				}).Debug("stored in cache")
			}
		}
	}

	return n
}

func (c *Memory) Get(q dns.Question) []dns.RR {
	if q.Qclass != dns.ClassINET {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	rrs, ok := c.data[key{
		Host: q.Name,
		Type: q.Qtype,
	}]
	if !ok || rrs == nil {
		return nil
	}

	return rrs.RR()
}
