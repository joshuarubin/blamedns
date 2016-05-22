package trie

import (
	"container/list"
)

type Node struct {
	C      byte
	Parent *Node
	Next   map[byte]*Node
	Value  interface{}
}

func (t *Node) Insert(text string, value interface{}) *Node {
	l := len(text)
	for i := 0; i <= l; i++ {
		if i == l {
			if value != nil {
				t.Value = value
			}
			return t
		}

		c := text[i]

		if t.Next == nil {
			t.Next = map[byte]*Node{}
		}

		n, ok := t.Next[c]
		if !ok {
			n = &Node{
				C:      c,
				Parent: t,
			}
			t.Next[c] = n
		}
		t = n
	}

	return nil
}

func (t *Node) Get(text string) *Node {
	return t.get(text, false)
}

func (t *Node) GetSubString(text string) *Node {
	return t.get(text, true)
}

func (t *Node) get(text string, matchSubString bool) *Node {
	l := len(text)
	for i := 0; i <= l; i++ {
		if (matchSubString && t.Value != nil) || i == l {
			if t.Value == nil {
				return nil
			}

			return t
		}

		c := text[i]
		var ok bool
		if t, ok = t.Next[c]; !ok {
			return nil
		}
	}

	return nil
}

type Walker interface {
	Walk(text string, node *Node)
}

type WalkerFunc func(text string, node *Node)

func (f WalkerFunc) Walk(text string, node *Node) {
	f(text, node)
}

type listNode struct {
	String   string
	TrieNode *Node
}

func (t *Node) listNode(str string) *listNode {
	return &listNode{
		String:   str,
		TrieNode: t,
	}
}

func (t *Node) Walk(w Walker) {
	queue := list.New()

	// initialize the queue
	for c, n := range t.Next {
		queue.PushBack(n.listNode(string(c)))
	}

	for queue.Len() > 0 {
		ln := queue.Remove(queue.Front()).(*listNode)

		w.Walk(ln.String, ln.TrieNode)

		for c, n := range ln.TrieNode.Next {
			queue.PushBack(n.listNode(ln.String + string(c)))
		}
	}
}

func (t *Node) Delete() {
	if t.Parent != nil {
		delete(t.Parent.Next, t.C)
	}
}

func (t *Node) DeletePrune() {
	t.Value = nil
	for ; t.Parent != nil; t = t.Parent {
		if t.Value != nil || len(t.Next) > 0 {
			break
		}

		t.Delete()
	}
}

func (t *Node) Prune() {
	t.Walk(WalkerFunc(func(text string, node *Node) {
		node.DeletePrune()
	}))
}

func (t *Node) Len() int {
	var n int
	t.Walk(WalkerFunc(func(text string, node *Node) {
		if node.Value != nil {
			n++
		}
	}))
	return n
}

func (t *Node) NumNodes() int {
	var n int
	t.Walk(WalkerFunc(func(text string, node *Node) {
		n++
	}))
	return n
}
