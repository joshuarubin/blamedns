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

func (c *Memory) add(rr dns.RR) *RR {
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
					"name": e.Header().Name,
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

	c.mu.Lock()
	defer c.mu.Unlock()

	k := key{
		Host: q.Name,
		Type: q.Qtype,
	}

	if elem, ok := c.data[k]; ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*entry).Set.RR()
	}

	return nil
}
