package blocker

import (
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/trie"
)

func addSource(t *trie.Node, source string) {
	if value := t.Value(); value != nil {
		t.Lock()
		value.(*sources).Add(source)
		t.Unlock()
		return
	}

	if t.SetValueIfNil(newSources(source)) {
		return
	}

	addSource(t, source)
}

func hasSource(t *trie.Node, source string) bool {
	if t == nil {
		return false
	}

	value := t.Value()

	if value == nil {
		return false
	}

	t.RLock()
	defer t.RUnlock()

	return value.(*sources).Has(source)
}

func numSources(t *trie.Node) int {
	if t == nil {
		return 0
	}

	value := t.Value()

	if value == nil {
		return 0
	}

	t.RLock()
	defer t.RUnlock()

	return value.(*sources).Len()
}

func deleteSource(t *trie.Node, source string) {
	if t == nil {
		return
	}

	value := t.Value()

	if value == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	value.(*sources).Remove(source)
}

func removeBySource(t *trie.Node, source string) {
	t.Walk(trie.WalkerFunc(func(node *trie.Node) {
		if !hasSource(node, source) {
			return
		}

		if numSources(node) > 1 {
			deleteSource(node, source)
			return
		}

		node.DeletePrune()
	}))
}

type TrieBlocker struct {
	trie trie.Node
}

func (b *TrieBlocker) AddHost(source, host string) {
	host = parser.ReverseHostName(host)
	n := b.trie.InsertString(host)
	addSource(n, source)
}

func (b *TrieBlocker) Block(host string) bool {
	host = parser.ReverseHostName(host)
	n := b.trie.GetString(host)
	if n == nil {
		return false
	}

	return n.Value() != nil
}

func (b *TrieBlocker) Len() int {
	return b.trie.Len()
}

func (b *TrieBlocker) Reset(source string) {
	removeBySource(&b.trie, source)
}
