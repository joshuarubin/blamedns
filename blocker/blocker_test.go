package blocker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBlocker(t *testing.T) {
	Convey("blockers should work", t, func() {
		for _, t := range []Blocker{&RadixBlocker{}, &HashBlocker{}, &SliceBlocker{}} {
			So(t.Len(), ShouldEqual, 0)

			t.AddHost("the source", "www.example.com")
			So(t.Len(), ShouldEqual, 1)
			So(t.Block("www.example.com"), ShouldBeTrue)

			t.AddHost("another source", "www.example.com")
			So(t.Len(), ShouldEqual, 1)
			So(t.Block("www.example.com"), ShouldBeTrue)

			So(t.Block("example.com"), ShouldBeFalse)
			So(t.Block("sub.www.example.com"), ShouldBeFalse)

			t.Reset("the source")
			So(t.Len(), ShouldEqual, 1)
			So(t.Block("www.example.com"), ShouldBeTrue)

			t.Reset("another source")
			So(t.Len(), ShouldEqual, 0)
			So(t.Block("www.example.com"), ShouldBeFalse)
		}
	})
}
