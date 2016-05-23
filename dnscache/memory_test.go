package dnscache

import (
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
	for _, sets := range c.data {
		n += len(*sets)
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

func TestMemoryCache(t *testing.T) {
	Convey("dns memory cache should work", t, func() {
		c := &Memory{Logger: logger}
		So(c.Len(), ShouldEqual, 0)

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
		c.FlushAll()
		So(c.Len(), ShouldEqual, 0)
		So(c.numEntries(), ShouldEqual, 0)
	})
}
