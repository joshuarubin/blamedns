package trie

import "container/list"

type Node struct {
	C      byte
	Parent *Node
	Next   map[byte]*Node
	Value  interface{}
}

func (t *Node) Insert(text string) *Node {
	l := len(text)
	for i := 0; i <= l; i++ {
		if i == l {
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

func (t *Node) Get(text string) interface{} {
	return t.get(text, false)
}

func (t *Node) GetSubString(text string) interface{} {
	return t.get(text, true)
}

func (t *Node) get(text string, matchSubString bool) interface{} {
	l := len(text)
	for i := 0; i <= l; i++ {
		if (matchSubString && t.Value != nil) || i == l {
			return t.Value
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
		e := queue.Front()
		queue.Remove(e)
		ln := e.Value.(*listNode)

		w.Walk(ln.String, ln.TrieNode)

		for c, n := range ln.TrieNode.Next {
			queue.PushBack(n.listNode(ln.String + string(c)))
		}
	}
}

func (t *Node) Delete() {
	delete(t.Parent.Next, t.C)
}

func (t *Node) Prune() {
	var n int
	t.Walk(WalkerFunc(func(text string, node *Node) {
		if node.Value == nil && len(node.Next) == 0 {
			n++
			node.Delete()
		}
	}))

	if n > 0 {
		t.Prune()
	}
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
