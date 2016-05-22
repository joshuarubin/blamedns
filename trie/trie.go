package trie

import (
	"bytes"
	"container/list"
	"sync"
)

type Node struct {
	sync.RWMutex
	c      byte
	parent *Node
	n      map[byte]*Node
	value  interface{}
}

func (t *Node) next(c byte) *Node {
	t.RLock()
	defer t.RUnlock()

	n, ok := t.n[c]
	if !ok {
		return nil
	}
	return n
}

func (t *Node) Value() interface{} {
	t.RLock()
	defer t.RUnlock()
	return t.value
}

func (t *Node) nextCreate(c byte) *Node {
	if n := t.next(c); n != nil {
		return n
	}

	t.Lock()
	defer t.Unlock()

	if t.n == nil {
		t.n = map[byte]*Node{}
	} else if n := t.n[c]; n != nil {
		// check to make sure that next wasn't set while upgrading the lock
		return n
	}

	n := &Node{
		c:      c,
		parent: t,
	}
	t.n[c] = n

	return n
}

func (t *Node) SetValue(value interface{}) {
	t.Lock()
	defer t.Unlock()
	t.value = value
}

func (t *Node) SetValueIfNil(value interface{}) bool {
	if t.Value() != nil {
		return false
	}

	t.Lock()
	defer t.Unlock()

	// make sure value wasn't set while upgrading the lock
	if t.value != nil {
		return false
	}

	t.value = value
	return true
}

func (t *Node) InsertString(text string) *Node {
	return t.Insert([]byte(text))
}

func (t *Node) Insert(text []byte) *Node {
	l := len(text)
	for i := 0; i <= l; i++ {
		if i == l {
			return t
		}

		t = t.nextCreate(text[i])
	}

	return nil
}

func (t *Node) GetString(text string) *Node {
	return t.Get([]byte(text))
}

func (t *Node) Get(text []byte) *Node {
	return t.get(text, false)
}

func (t *Node) GetSubString(text string) *Node {
	return t.GetSub([]byte(text))
}

func (t *Node) GetSub(text []byte) *Node {
	return t.get(text, true)
}

func (t *Node) get(text []byte, matchSubString bool) *Node {
	l := len(text)
	for i := 0; i <= l; i++ {
		value := t.Value()

		if (matchSubString && value != nil) || i == l {
			if value == nil {
				return nil
			}

			return t
		}

		if t = t.next(text[i]); t == nil {
			return nil
		}
	}

	return nil
}

type Walker interface {
	Walk(node *Node)
}

type WalkerFunc func(node *Node)

func (f WalkerFunc) Walk(node *Node) {
	f(node)
}

func (t *Node) pushNext(queue *list.List) {
	t.RLock()
	defer t.RUnlock()

	for _, n := range t.n {
		if n != nil {
			queue.PushBack(n)
		}
	}
}

func pop(queue *list.List) *Node {
	if queue.Len() == 0 {
		return nil
	}

	return queue.Remove(queue.Front()).(*Node)
}

func (t *Node) Walk(w Walker) {
	queue := list.New()

	t.pushNext(queue)

	for ln := pop(queue); ln != nil; ln = pop(queue) {
		w.Walk(ln)
		ln.pushNext(queue)
	}
}

func (t *Node) deleteChild(c byte) {
	t.Lock()
	defer t.Unlock()
	delete(t.n, c)
}

func (t *Node) Delete() {
	if t.parent != nil {
		t.parent.deleteChild(t.c)
	}
}

func (t *Node) lenNext() int {
	t.RLock()
	defer t.RUnlock()

	var ret int
	for _, n := range t.n {
		if n != nil {
			ret++
		}
	}

	return ret
}

func (t *Node) DeletePrune() {
	t.SetValue(nil)
	for ; t.parent != nil; t = t.parent {
		if t.Value() != nil || t.lenNext() > 0 {
			break
		}

		t.Delete()
	}
}

func (t *Node) Key() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 32))

	for ; t.parent != nil; t = t.parent {
		if err := buf.WriteByte(t.c); err != nil {
			panic(err)
		}
	}

	// the buffer needs to be reversed before it can be returned
	ret := buf.Bytes()
	l := len(ret)

	for i := 0; i < l/2; i++ {
		j := l - 1 - i
		ret[j], ret[i] = ret[i], ret[j]
	}

	return ret
}

func (t *Node) KeyString() string {
	return string(t.Key())
}

func (t *Node) Prune() {
	t.Walk(WalkerFunc(func(node *Node) {
		node.DeletePrune()
	}))
}

func (t *Node) Len() int {
	var n int
	t.Walk(WalkerFunc(func(node *Node) {
		if node.Value() != nil {
			n++
		}
	}))
	return n
}

func (t *Node) NumNodes() int {
	var n int
	t.Walk(WalkerFunc(func(node *Node) {
		n++
	}))
	return n
}
