package dnscachetypes

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRRSet(t *testing.T) {
	Convey("rrset should work", t, func() {
		rrs := RRSet{}

		So(rrs.Len(), ShouldEqual, 0)

		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name:   "example.com",
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: net.ParseIP("1.2.3.4"),
		}

		ret := rrs.Add(rr)
		So(ret, ShouldNotBeNil)
		So(rrs.Len(), ShouldEqual, 1)
		So(ret.Expires.After(time.Now().UTC().Add(59*time.Second)), ShouldBeTrue)
		So(ret.Expires.After(time.Now().UTC().Add(61*time.Second)), ShouldBeFalse)
		r := rrs.RR()
		So(r, ShouldNotBeNil)
		So(len(r), ShouldEqual, 1)

		rr.Hdr.Ttl = 1
		ret = rrs.Add(rr)
		So(ret, ShouldNotBeNil)
		So(rrs.Len(), ShouldEqual, 1)
		So(ret.Expires.After(time.Now().UTC()), ShouldBeTrue)
		So(ret.Expires.After(time.Now().UTC().Add(2*time.Second)), ShouldBeFalse)
		r = rrs.RR()
		So(r, ShouldNotBeNil)
		So(len(r), ShouldEqual, 1)

		time.Sleep(1 * time.Second)

		rr.Hdr.Ttl = 60
		ret = rrs.Add(rr)
		So(ret, ShouldNotBeNil)
		So(rrs.Len(), ShouldEqual, 1)
		So(ret.Expires.After(time.Now().UTC().Add(59*time.Second)), ShouldBeTrue)
		So(ret.Expires.After(time.Now().UTC().Add(61*time.Second)), ShouldBeFalse)
		r = rrs.RR()
		So(r, ShouldNotBeNil)
		So(len(r), ShouldEqual, 1)

		rr.Hdr.Name = "www.example.com"
		rr.Hdr.Ttl = 1
		ret = rrs.Add(rr)
		So(ret, ShouldNotBeNil)
		So(rrs.Len(), ShouldEqual, 2)
		So(ret.Expires.After(time.Now().UTC()), ShouldBeTrue)
		So(ret.Expires.After(time.Now().UTC().Add(2*time.Second)), ShouldBeFalse)
		r = rrs.RR()
		So(r, ShouldNotBeNil)
		So(len(r), ShouldEqual, 2)

		time.Sleep(1 * time.Second)
		So(rrs.Prune(), ShouldEqual, 1)
		So(rrs.Len(), ShouldEqual, 1)

		rr.Hdr.Name = "example.com"
		rr.Hdr.Ttl = 60
		rr.Hdr.Rrtype = dns.TypeCNAME
		ret = rrs.Add(rr)
		So(ret, ShouldNotBeNil)
		So(rrs.Len(), ShouldEqual, 2)
		So(ret.Expires.After(time.Now().UTC().Add(59*time.Second)), ShouldBeTrue)
		So(ret.Expires.After(time.Now().UTC().Add(61*time.Second)), ShouldBeFalse)
		r = rrs.RR()
		So(r, ShouldNotBeNil)
		So(len(r), ShouldEqual, 2)
	})
}
