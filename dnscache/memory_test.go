package dnscache

import (
	"fmt"
	"net"
	"testing"
	"time"

	"jrubin.io/blamedns/lru"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func (c *Memory) numEntries() int {
	var n int
	c.data.Each(lru.ElementerFunc(func(k, value interface{}) {
		n += len(*value.(*RRSet))
	}))
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
				Target: "example.com",
			},
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.1"),
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

	msg1 = &dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "a.example.com",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Target: "b.example.com",
			},
		},
	}

	// intentionally make a cname loop with msg1
	msg2 = &dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "b.example.com",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Target: "c.example.com",
			},
		},
	}

	// intentionally make a cname loop with msg1
	msg3 = &dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "c.example.com",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Target: "a.example.com",
			},
		},
	}

	logger = logrus.New()
)

func init() {
	logger.Level = logrus.DebugLevel
}

func newAMsg(host string) *dns.Msg {
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
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.1"),
			},
		},
	}
}

func newReq(typ uint16, host string) *dns.Msg {
	return &dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
			Name:   host,
			Qtype:  typ,
		}},
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

		So(c.Set(&dns.Msg{Question: []dns.Question{{}}}), ShouldEqual, 0)
		So(c.Len(), ShouldEqual, 0)

		So(msg, ShouldNotBeNil)

		So(c.Set(msg), ShouldEqual, 5)
		So(c.Len(), ShouldEqual, 5)
		So(c.numEntries(), ShouldEqual, 5)

		resp := c.Get(nil, newReq(dns.TypeA, "example.com"))

		So(resp, ShouldNotBeNil)
		So(resp.Answer, ShouldNotBeNil)
		So(len(resp.Answer), ShouldEqual, 2)
		for i := range resp.Answer {
			r, ok := resp.Answer[i].(*dns.A)
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

		resp = c.Get(nil, newReq(dns.TypeA, "cname.example.com"))
		So(resp, ShouldNotBeNil)
		So(resp.Answer, ShouldNotBeNil)
		So(len(resp.Answer), ShouldEqual, 3)

		resp = c.Get(nil, newReq(dns.TypeAAAA, "cname.example.com"))
		So(resp, ShouldNotBeNil)
		So(resp.Answer, ShouldNotBeNil)
		So(len(resp.Answer), ShouldEqual, 2)

		resp = c.Get(nil, newReq(dns.TypeAAAA, "example.com"))
		So(resp, ShouldNotBeNil)
		So(resp.Answer, ShouldNotBeNil)
		So(len(resp.Answer), ShouldEqual, 1)

		r, ok := resp.Answer[0].(*dns.AAAA)
		So(ok, ShouldBeTrue)
		So(r.Hdr.Name, ShouldEqual, "example.com")
		So(r.Hdr.Rrtype, ShouldEqual, dns.TypeAAAA)
		So(r.Hdr.Class, ShouldEqual, dns.ClassINET)
		So(r.Hdr.Ttl, ShouldEqual, 0)
		So(r.AAAA.String(), ShouldEqual, "::1")

		time.Sleep(1 * time.Second)

		resp = c.Get(nil, newReq(dns.TypeAAAA, "example.com"))
		So(resp, ShouldBeNil)
		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)

		So(c.Prune(), ShouldEqual, 0)
		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)

		req := newReq(dns.TypeA, "example.com")
		req.Question[0].Qclass = 0 // missing Qclass
		resp = c.Get(nil, req)
		So(resp, ShouldBeNil)

		resp = c.Get(nil, newReq(dns.TypeA, "www.example.com"))
		So(resp, ShouldBeNil)

		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)
		c.Purge()
		So(c.Len(), ShouldEqual, 0)
		So(c.numEntries(), ShouldEqual, 0)
	})

	Convey("test lru", t, func() {
		Convey("lru should work", func() {
			evictCounter := 0
			onEvicted := lru.ElementerFunc(func(k, value interface{}) {
				typ := k.(key).Type
				host := k.(key).Host
				rr := value.(*RRSet).RR()

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
				l.Set(newAMsg(fmt.Sprintf("%d.example.com", i)))
			}

			So(l.Len(), ShouldEqual, 128)
			So(evictCounter, ShouldEqual, 128)

			for i := 0; i < 128; i++ {
				rr := l.Get(nil, newReq(dns.TypeA, fmt.Sprintf("%d.example.com", i+128)))
				So(rr, ShouldNotBeNil)
			}

			for i := 0; i < 128; i++ {
				rr := l.Get(nil, newReq(dns.TypeA, fmt.Sprintf("%d.example.com", i)))
				So(rr, ShouldBeNil)
			}

			l.Purge()
			So(l.Len(), ShouldEqual, 0)
		})

		Convey("lru add", func() {
			// Test that Add returns true/false if an eviction occurred
			evictCounter := 0
			onEvicted := lru.ElementerFunc(func(k, value interface{}) {
				evictCounter++
			})

			l := NewMemory(1, nil, onEvicted)

			l.Set(newAMsg("0.example.com"))
			So(evictCounter, ShouldEqual, 0)

			l.Set(newAMsg("1.example.com"))
			So(evictCounter, ShouldEqual, 1)
		})

		Convey("cname loops shouldnt kill the server", func() {
			c := NewMemory(128, nil, nil)
			// add a cname record
			So(c.Set(msg1), ShouldEqual, 1)

			// do an A lookup for it
			r := testGet(c, dns.TypeA, "a.example.com")
			So(r, ShouldBeNil)

			// now add a cname that points to a third record
			So(c.Set(msg2), ShouldEqual, 1)
			r = testGet(c, dns.TypeA, "b.example.com")
			So(r, ShouldBeNil)

			// getting the first record should return both
			r = testGet(c, dns.TypeA, "a.example.com")
			So(r, ShouldBeNil)

			// finally add a cname that points back to the first record, making
			// a cname loop
			So(c.Set(msg3), ShouldEqual, 1)
			r = testGet(c, dns.TypeA, "c.example.com")
			So(r, ShouldBeNil)
		})
	})
}

func testGet(c Cache, t uint16, host string) *dns.Msg {
	return c.Get(nil, &dns.Msg{
		Question: []dns.Question{{
			Name:   host,
			Qtype:  t,
			Qclass: dns.ClassINET,
		}},
	})
}
