package blocker

import (
	"testing"

	"jrubin.io/blamedns/trie"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrie(t *testing.T) {
	Convey("reverse should work", t, func() {
		So(ReverseHostName(""), ShouldEqual, "")
		So(ReverseHostName("com"), ShouldEqual, "com")
		So(ReverseHostName("example.com"), ShouldEqual, "com.example")
		So(ReverseHostName("www.example.com"), ShouldEqual, "com.example.www")
		So(ReverseHostName("sub.www.example.com"), ShouldEqual, "com.example.www.sub")
		So(ReverseHostName("example..com"), ShouldEqual, "com..example")
	})

	Convey("trie should work", t, func() {
		t := trie.Node{}
		n := t.Insert(ReverseHostName("www.example.com"), nil)
		addSource(n, "the source")

		s := t.Get(ReverseHostName("www.example.com"))
		So(s, ShouldNotBeNil)
		So(s.Value, ShouldNotBeNil)
		So(numSources(s), ShouldEqual, 1)
		So(hasSource(s, "the source"), ShouldBeTrue)

		n = t.Insert(ReverseHostName("www.example.com"), nil)
		addSource(n, "another source")

		s = t.Get(ReverseHostName("www.example.com"))
		So(s, ShouldNotBeNil)
		So(s.Value, ShouldNotBeNil)
		So(numSources(s), ShouldEqual, 2)
		So(hasSource(s, "the source"), ShouldBeTrue)
		So(hasSource(s, "another source"), ShouldBeTrue)

		s = t.Get(ReverseHostName("example.com"))
		So(s, ShouldBeNil)

		s = t.Get(ReverseHostName("sub.www.example.com"))
		So(s, ShouldBeNil)

		s = t.GetSubString(ReverseHostName("www.example.com"))
		So(s, ShouldNotBeNil)
		So(s.Value, ShouldNotBeNil)
		So(numSources(s), ShouldEqual, 2)
		So(hasSource(s, "the source"), ShouldBeTrue)
		So(hasSource(s, "another source"), ShouldBeTrue)

		s = t.GetSubString(ReverseHostName("example.com"))
		So(s, ShouldBeNil)

		s = t.GetSubString(ReverseHostName("ww.example.com"))
		So(s, ShouldBeNil)

		s = t.GetSubString(ReverseHostName("sub.www.example.com"))
		So(s, ShouldNotBeNil)
		So(s.Value, ShouldNotBeNil)
		So(numSources(s), ShouldEqual, 2)
		So(hasSource(s, "the source"), ShouldBeTrue)
		So(hasSource(s, "another source"), ShouldBeTrue)

		s = t.GetSubString(ReverseHostName("www2.example.com"))
		So(s, ShouldNotBeNil)
		So(s.Value, ShouldNotBeNil)
		So(numSources(s), ShouldEqual, 2)
		So(hasSource(s, "the source"), ShouldBeTrue)
		So(hasSource(s, "another source"), ShouldBeTrue)

		So(t.Len(), ShouldEqual, 1)
		So(t.NumNodes(), ShouldEqual, len("www.example.com"))

		removeBySource(&t, "the source")

		s = t.Get(ReverseHostName("www.example.com"))
		So(s, ShouldNotBeNil)
		So(s.Value, ShouldNotBeNil)
		So(numSources(s), ShouldEqual, 1)
		So(hasSource(s, "another source"), ShouldBeTrue)
		So(t.Len(), ShouldEqual, 1)
		So(t.NumNodes(), ShouldEqual, len("www.example.com"))

		removeBySource(&t, "another source")

		s = t.Get(ReverseHostName("www.example.com"))
		So(s, ShouldBeNil)
		So(t.Len(), ShouldEqual, 0)
		So(t.NumNodes(), ShouldEqual, 0)
	})
}
