package trie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrie(t *testing.T) {
	Convey("trie should work", t, func() {
		t := Node{}
		n := t.InsertString("www.example.com")
		n.SetValue("the source")

		n = t.GetString("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value(), ShouldNotBeNil)
		So(n.Value(), ShouldEqual, "the source")

		n = t.InsertString("www.example.com")
		n.SetValue("another source")

		n = t.GetString("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value(), ShouldNotBeNil)
		So(n.Value(), ShouldEqual, "another source")

		n = t.GetString("example.com")
		So(n, ShouldBeNil)

		n = t.GetString("sub.www.example.com")
		So(n, ShouldBeNil)

		n = t.GetSubString("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value(), ShouldNotBeNil)
		So(n.Value(), ShouldEqual, "another source")

		n = t.GetSubString("example.com")
		So(n, ShouldBeNil)

		n = t.GetSubString("ww.example.com")
		So(n, ShouldBeNil)

		n = t.GetSubString("www.example.com.sub")
		So(n, ShouldNotBeNil)
		So(n.Value(), ShouldNotBeNil)
		So(n.Value(), ShouldEqual, "another source")

		n = t.GetSubString("www.example.com2")
		So(n, ShouldNotBeNil)
		So(n.Value(), ShouldNotBeNil)
		So(n.Value(), ShouldEqual, "another source")

		So(t.Len(), ShouldEqual, 1)
		So(t.NumNodes(), ShouldEqual, len("www.example.com"))

		n = t.GetString("www.example.com")
		So(n, ShouldNotBeNil)
		So(n.Value(), ShouldNotBeNil)
		So(n.Value(), ShouldEqual, "another source")

		n.Delete()
		So(t.Len(), ShouldEqual, 0)
		So(t.NumNodes(), ShouldEqual, len("www.example.com")-1)

		t.Prune()
		So(t.Len(), ShouldEqual, 0)
		So(t.NumNodes(), ShouldEqual, 0)

		n = t.InsertString("abc")
		So(n.KeyString(), ShouldEqual, "abc")
		n.SetValue(true)
		So(t.Len(), ShouldEqual, 1)
		So(t.NumNodes(), ShouldEqual, len("abc"))

		n = t.GetString("ab")
		So(n, ShouldBeNil)

		n = t.GetSubString("ab")
		So(n, ShouldBeNil)

		n = t.InsertString("abd")
		So(n.KeyString(), ShouldEqual, "abd")
		So(n.SetValueIfNil(true), ShouldBeTrue)
		So(t.Len(), ShouldEqual, 2)
		So(t.NumNodes(), ShouldEqual, len("abc")+1)

		So(n.SetValueIfNil(true), ShouldBeFalse)
	})
}
