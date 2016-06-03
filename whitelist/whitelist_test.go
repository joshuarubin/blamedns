package whitelist

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWhiteList(t *testing.T) {
	Convey("whitelist should work", t, func() {
		wl := New("example.com", "www.example.com")
		So(wl, ShouldNotBeNil)
		So(wl.Pass("example.com"), ShouldBeTrue)
		So(wl.Pass("www.example.com"), ShouldBeTrue)
		So(wl.Pass("com"), ShouldBeFalse)
		So(wl.Pass("srv.www.example.com"), ShouldBeFalse)
	})
}
