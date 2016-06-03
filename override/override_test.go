package override

import (
	"net"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func ipsEqual(in []net.IP, vals ...string) bool {
	if len(in) != len(vals) {
		return false
	}

	for i, ip := range in {
		if !ip.Equal(net.ParseIP(vals[i])) {
			return false
		}
	}

	return true
}

func TestOverride(t *testing.T) {
	Convey("override should work", t, func() {
		ips := ParseIPs("127.0.0.1", "127.0.0.2")
		So(ipsEqual(ips, "127.0.0.1", "127.0.0.2"), ShouldBeTrue)

		list := map[string][]net.IP{
			"example.com": ips,
		}

		o := New(list)
		So(o, ShouldNotBeNil)

		v := o.Override("www.example.com")
		So(v, ShouldBeNil)

		v = o.Override("example.com")
		So(ipsEqual(v, "127.0.0.1", "127.0.0.2"), ShouldBeTrue)
	})
}
