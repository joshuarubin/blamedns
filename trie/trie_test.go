package trie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrie(t *testing.T) {
	Convey("trie should work", t, func() {
		t := Node{}
		t.Insert("www.example.com", "the source")

		n := t.Get("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value, ShouldNotBeNil)
		So(n.Value, ShouldEqual, "the source")

		t.Insert("www.example.com", "another source")

		n = t.Get("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value, ShouldNotBeNil)
		So(n.Value, ShouldEqual, "another source")

		n = t.Get("example.com")
		So(n, ShouldBeNil)

		n = t.Get("sub.www.example.com")
		So(n, ShouldBeNil)

		n = t.GetSubString("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value, ShouldNotBeNil)
		So(n.Value, ShouldEqual, "another source")

		n = t.GetSubString("example.com")
		So(n, ShouldBeNil)

		n = t.GetSubString("ww.example.com")
		So(n, ShouldBeNil)

		n = t.GetSubString("www.example.com.sub")
		So(n, ShouldNotBeNil)
		So(n.Value, ShouldNotBeNil)
		So(n.Value, ShouldEqual, "another source")

		n = t.GetSubString("www.example.com2")
		So(n, ShouldNotBeNil)
		So(n.Value, ShouldNotBeNil)
		So(n.Value, ShouldEqual, "another source")

		So(t.Len(), ShouldEqual, 1)
		So(t.NumNodes(), ShouldEqual, len("www.example.com"))

		n = t.Get("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value, ShouldNotBeNil)
		So(n.Value, ShouldEqual, "another source")

		n.Delete()
		So(t.Len(), ShouldEqual, 0)
		So(t.NumNodes(), ShouldEqual, len("www.example.com")-1)

		t.Prune()
		So(t.Len(), ShouldEqual, 0)
		So(t.NumNodes(), ShouldEqual, 0)

		t.Insert("abc", true)
		So(t.Len(), ShouldEqual, 1)
		So(t.NumNodes(), ShouldEqual, len("abc"))

		n = t.Get("ab")
		So(n, ShouldBeNil)
	})
}
