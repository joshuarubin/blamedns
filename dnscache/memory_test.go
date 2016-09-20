package dnscache

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"

	"github.com/miekg/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func (c *Memory) numEntries() int {
	var n int
	for _, key := range c.cache.Keys() {
		value, ok := c.cache.Get(key)
		if !ok {
			continue
		}

		switch v := value.(type) {
		case *RRSet:
			n += len(v.data)
		case *negativeEntry:
			n++
		default:
			panic(fmt.Errorf("invalid type stored in dns memory cache"))
		}
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

	logger = text.Logger(slog.DebugLevel)
)

func TestMemoryCache(t *testing.T) {
	Convey("dns memory cache should work", t, func() {
		So(func() { NewMemory(0, nil) }, ShouldPanic)

		c := NewMemory(1, nil)
		So(c.Logger, ShouldNotEqual, logger)

		c = NewMemory(64, logger)
		So(c.Len(), ShouldEqual, 0)
		So(c.Logger, ShouldEqual, logger)

		So(c.Set(&dns.Msg{Question: []dns.Question{{}}}), ShouldEqual, 0)
		So(c.Len(), ShouldEqual, 0)

		So(msg, ShouldNotBeNil)

		So(c.Set(msg), ShouldEqual, 5)
		So(c.Len(), ShouldEqual, 5)
		So(c.numEntries(), ShouldEqual, 5)

		resp := testGet(c, dns.TypeA, "example.com")

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

		resp = testGet(c, dns.TypeA, "cname.example.com")
		So(resp, ShouldNotBeNil)
		So(resp.Answer, ShouldNotBeNil)
		So(len(resp.Answer), ShouldEqual, 3)

		resp = testGet(c, dns.TypeAAAA, "cname.example.com")
		So(resp, ShouldNotBeNil)
		So(resp.Answer, ShouldNotBeNil)
		So(len(resp.Answer), ShouldEqual, 2)

		resp = testGet(c, dns.TypeAAAA, "example.com")
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

		resp = testGet(c, dns.TypeAAAA, "example.com")
		So(resp, ShouldBeNil)
		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)

		So(c.Prune(), ShouldEqual, 0)
		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)

		req := testGet(c, dns.TypeA, "example.com")
		req.Question[0].Qclass = 0 // missing Qclass
		resp = c.Get(nil, req)
		So(resp, ShouldBeNil)

		resp = testGet(c, dns.TypeA, "www.example.com")
		So(resp, ShouldBeNil)

		So(c.Len(), ShouldEqual, 4)
		So(c.numEntries(), ShouldEqual, 4)
		c.Purge()
		So(c.Len(), ShouldEqual, 0)
		So(c.numEntries(), ShouldEqual, 0)
	})

	Convey("test lru", t, func() {
		Convey("lru should work", func() {
			l := NewMemory(128, nil)

			for i := 0; i < 256; i++ {
				testSetA(l, fmt.Sprintf("%d.example.com", i), nil, 0)
			}

			So(l.Len(), ShouldEqual, 128)

			for i := 0; i < 128; i++ {
				rr := testGet(l, dns.TypeA, fmt.Sprintf("%d.example.com", i+128))
				So(rr, ShouldNotBeNil)
			}

			for i := 0; i < 128; i++ {
				rr := testGet(l, dns.TypeA, fmt.Sprintf("%d.example.com", i))
				So(rr, ShouldBeNil)
			}

			l.Purge()
			So(l.Len(), ShouldEqual, 0)
		})

		Convey("cname loops shouldnt kill the server", func() {
			c := NewMemory(128, nil)
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

	Convey("memory cache should not have any races", t, func() {
		// start 4 goroutines, 2 setting values and 2 getting values
		n := 1024
		c := NewMemory(n, nil)
		c.Start(1 * time.Second) // prune every second
		hosts := genHosts(n)
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		for i := 0; i < 2; i++ {
			go fastSetter(ctx, c, hosts)
			go fastGetter(ctx, c, hosts)
		}
		<-ctx.Done()
	})
}

func genHosts(n int) []string {
	ret := make([]string, n)
	for i := 0; i < n; i++ {
		ret[i] = fmt.Sprintf("%d.example.com", i)
	}
	return ret
}

func getRandInt(max int) int {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(i.Int64())
}

func fastSetter(ctx context.Context, c Cache, hosts []string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		i := getRandInt(len(hosts))
		ttl := TTL(getRandInt(int(5 * time.Second)))

		testSetA(c, hosts[i], nil, ttl)
	}
}

func fastGetter(ctx context.Context, c Cache, hosts []string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		i := getRandInt(len(hosts))
		testGet(c, dns.TypeA, hosts[i])
	}
}

func testSetA(c Cache, host string, ip net.IP, ttl TTL) {
	if ip == nil {
		ip = net.ParseIP("127.0.0.1")
	}

	if ttl.Seconds() < 1 {
		ttl = TTL(60 * time.Second)
	}

	c.Set(&dns.Msg{
		Question: []dns.Question{{
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   host,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    ttl.Seconds(),
				},
				A: ip,
			},
		},
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
