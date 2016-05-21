package blocker

import "jrubin.io/blamedns/trie"

func addSource(t *trie.Node, source string) {
	if t.Value == nil {
		t.Value = map[string]struct{}{}
	}

	value := t.Value.(map[string]struct{})
	value[source] = struct{}{}
}

func hasSource(t interface{}, source string) bool {
	if t == nil {
		return false
	}

	value := t.(map[string]struct{})
	_, ok := value[source]
	return ok
}

func numSources(t interface{}) int {
	if t == nil {
		return 0
	}

	return len(t.(map[string]struct{}))
}

func deleteSource(t interface{}, source string) {
	if t == nil {
		return
	}

	value := t.(map[string]struct{})
	delete(value, source)
}

func removeBySource(t *trie.Node, source string) {
	var n int
	t.Walk(trie.WalkerFunc(func(text string, node *trie.Node) {
		if hasSource(node.Value, source) {
			n++
			if numSources(node.Value) == 1 {
				node.Delete()
				return
			}

			deleteSource(node.Value, source)
		}
	}))

	if n > 0 {
		t.Prune()
	}
}
