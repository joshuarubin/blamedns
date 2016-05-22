package parser

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReverse(t *testing.T) {
	Convey("reverse should work", t, func() {
		So(ReverseHostName(""), ShouldEqual, "")
		So(ReverseHostName("com"), ShouldEqual, "com")
		So(ReverseHostName("example.com"), ShouldEqual, "com.example")
		So(ReverseHostName("www.example.com"), ShouldEqual, "com.example.www")
		So(ReverseHostName("sub.www.example.com"), ShouldEqual, "com.example.www.sub")
		So(ReverseHostName("example..com"), ShouldEqual, "com..example")
	})
}
