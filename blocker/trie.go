package blocker

import "jrubin.io/blamedns/trie"

func addSource(t *trie.Node, source string) {
	if t.Value == nil {
		t.Value = map[string]struct{}{}
	}

	value := t.Value.(map[string]struct{})
	value[source] = struct{}{}
}

func hasSource(t *trie.Node, source string) bool {
	if t == nil || t.Value == nil {
		return false
	}

	value := t.Value.(map[string]struct{})
	_, ok := value[source]
	return ok
}

func numSources(t *trie.Node) int {
	if t == nil || t.Value == nil {
		return 0
	}

	return len(t.Value.(map[string]struct{}))
}

func deleteSource(t *trie.Node, source string) {
	if t == nil || t.Value == nil {
		return
	}

	value := t.Value.(map[string]struct{})
	delete(value, source)
}

func removeBySource(t *trie.Node, source string) {
	t.Walk(trie.WalkerFunc(func(text string, node *trie.Node) {
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
