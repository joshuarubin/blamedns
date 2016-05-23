package dnscache

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func (c *Memory) numEntries() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var n int
	for _, elem := range c.data {
		n += len(*elem.Value.(*entry).Set)
	}
	return n
}

var (
	_ Cache = &Memory{}

	msg = &dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "cname.example.com",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Target: "example.com.",
			},
		},
		Ns: []dns.RR{
			&dns.SOA{
				Hdr: dns.RR_Header{
					Name:   "example.com",
					Rrtype: dns.TypeSOA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Ns:      "ns0.example.com",
				Mbox:    "ns.example.com",
				Serial:  1,
				Refresh: 1,
				Retry:   1,
				Expire:  1,
				Minttl:  1,
			},
		},
		Extra: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.1"),
			},
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.2"),
			},
			&dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   "example.com",
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    1,
				},
				AAAA: net.ParseIP("::1"),
			},
			&dns.A{},
		},
	}

	logger = logrus.New()
)

func init() {
	logger.Level = logrus.DebugLevel
}

func newAMsg(host string, ttl uint32, ip net.IP) *dns.Msg {
	return &dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   host,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    ttl,
				},
				A: ip,
			},
		},
	}
}

func newAQ(host string) dns.Question {
	return dns.Question{
		Qclass: dns.ClassINET,
		Name:   host,
		Qtype:  dns.TypeA,
	}
}

func TestMemoryCache(t *testing.T) {
	Convey("dns memory cache should work", t, func() {
		So(func() { NewMemory(0, nil, nil) }, ShouldPanic)

		c := NewMemory(1, nil, nil)
		So(c.Logger, ShouldNotEqual, logger)

		c = NewMemory(64, logger, nil)
		So(c.Len(), ShouldEqual, 0)
		So(c.Logger, ShouldEqual, logger)

		So(c.Set(&dns.Msg{}), ShouldEqual, 0)
		So(c.Len(), ShouldEqual, 0)

		So(c.Set(&dns.Msg{Question: []dns.Question{{}}}), ShouldEqual, 0)
		So(c.Len(), ShouldEqual, 0)

		So(msg, ShouldNotBeNil)

		So(c.Set(msg), ShouldEqual, 5)
		So(c.Len(), ShouldEqual, 5)
		So(c.numEntries(), ShouldEqual, 5)

		rr := c.Get(dns.Question{
			Qclass: dns.ClassINET,
			Name:   "example.com",
			Qtype:  dns.TypeA,
		})

		So(rr, ShouldNotBeNil)
		So(len(rr), ShouldEqual, 2)
		for i := range rr {
			r, ok := rr[i].(*dns.A)
			So(ok, ShouldBeTrue)
			So(r.Hdr.Name, ShouldEqual, "example.com")
			So(r.Hdr.Rrtype, ShouldEqual, dns.TypeA)
			So(r.Hdr.Class, ShouldEqual, dns.ClassINET)
			So(r.Hdr.Ttl, ShouldEqual, 59)

			if i == 0 {
				So(r.A.String(), ShouldEqual, "127.0.0.1")
			} else if i == 1 {
				So(r.A.String(), ShouldEqual, "127.0.0.2")
			}
		}

		rr = c.Get(dns.Question{
			Qclass: dns.ClassINET,
			Name:   "example.com",
			Qtype:  dns.TypeAAAA,
		})
		So(rr, ShouldNotBeNil)
		So(len(rr), ShouldEqual, 1)

		r, ok := rr[0].(*dns.AAAA)
		So(ok, ShouldBeTrue)
		So(r.Hdr.Name, ShouldEqual, "example.com")
		So(r.Hdr.Rrtype, ShouldEqual, dns.TypeAAAA)
		So(r.Hdr.Class, ShouldEqual, dns.ClassINET)
		So(r.Hdr.Ttl, ShouldEqual, 0)
		So(r.AAAA.String(), ShouldEqual, "::1")

		time.Sleep(1 * time.Second)

		rr = c.Get(dns.Question{
			Qclass: dns.ClassINET,
			Name:   "example.com",
			Qtype:  dns.TypeAAAA,
		})
		So(rr, ShouldBeNil)
		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 5)

		So(c.Prune(), ShouldEqual, 1)
		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)

		rr = c.Get(dns.Question{
			// missing Qclass
			Name:  "example.com",
			Qtype: dns.TypeA,
		})
		So(rr, ShouldBeNil)

		rr = c.Get(dns.Question{
			Qclass: dns.ClassINET,
			Name:   "www.example.com",
			Qtype:  dns.TypeA,
		})
		So(rr, ShouldBeNil)

		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)
		c.Purge()
		So(c.Len(), ShouldEqual, 0)
		So(c.numEntries(), ShouldEqual, 0)
	})

	Convey("test lru", t, func() {
		Convey("lru should work", func() {
			evictCounter := 0
			onEvicted := EvictNotifierFunc(func(typ uint16, host string, rr []dns.RR) {
				if evictCounter < 128 {
					So(typ, ShouldEqual, dns.TypeA)
					So(host, ShouldEqual, fmt.Sprintf("%d.example.com", evictCounter))
					So(len(rr), ShouldEqual, 1)
					So(rr[0], ShouldResemble, &dns.A{
						Hdr: dns.RR_Header{
							Name:   fmt.Sprintf("%d.example.com", evictCounter),
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    59,
						},
						A: net.ParseIP("127.0.0.1"),
					})
				}
				evictCounter++
			})

			l := NewMemory(128, nil, onEvicted)

			for i := 0; i < 256; i++ {
				l.Set(newAMsg(fmt.Sprintf("%d.example.com", i), 60, net.ParseIP("127.0.0.1")))
			}

			So(l.Len(), ShouldEqual, 128)
			So(evictCounter, ShouldEqual, 128)

			for i := 0; i < 128; i++ {
				rr := l.Get(newAQ(fmt.Sprintf("%d.example.com", i+128)))
				So(rr, ShouldNotBeNil)
			}

			for i := 0; i < 128; i++ {
				rr := l.Get(newAQ(fmt.Sprintf("%d.example.com", i)))
				So(rr, ShouldBeNil)
			}

			l.Purge()
			So(l.Len(), ShouldEqual, 0)
		})

		Convey("lru add", func() {
			// Test that Add returns true/false if an eviction occured
			evictCounter := 0
			onEvicted := EvictNotifierFunc(func(typ uint16, host string, rr []dns.RR) {
				evictCounter++
			})

			l := NewMemory(1, nil, onEvicted)

			l.Set(newAMsg("0.example.com", 60, net.ParseIP("127.0.0.1")))
			So(evictCounter, ShouldEqual, 0)

			l.Set(newAMsg("1.example.com", 60, net.ParseIP("127.0.0.1")))
			So(evictCounter, ShouldEqual, 1)
		})
	})
}
