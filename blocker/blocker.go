package blocker

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/trie"
)

var (
	_ parser.HostAdder  = &Blocker{}
	_ dnsserver.Blocker = &Blocker{}
)

type Blocker struct {
	Logger *logrus.Logger
	trie   trie.Node
	mu     sync.RWMutex
}

func (b *Blocker) AddHost(source, host string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.trie.Insert(ReverseHostName(host), nil)
	addSource(n, source)
}

func (b *Blocker) Block(host string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.trie.GetSubString(ReverseHostName(host)) != nil
}

func (b *Blocker) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.trie.Len()
}

func (b *Blocker) Reset(source string) {
	removeBySource(&b.trie, source)
}
