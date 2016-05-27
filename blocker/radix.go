package blocker

import (
	"sync"

	"jrubin.io/blamedns/parser"

	"github.com/armon/go-radix"
)

type RadixBlocker struct {
	data *radix.Tree
	mu   sync.RWMutex
}

func (b *RadixBlocker) AddHost(source, host string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := parser.ReverseHostName(host)

	if b.data == nil {
		b.data = radix.New()
		b.data.Insert(key, newSources(source))
		return
	}

	value, ok := b.data.Get(key)
	if !ok || value == nil {
		b.data.Insert(key, newSources(source))
		return
	}

	s, ok := value.(*sources)
	if !ok {
		b.data.Insert(key, newSources(source))
		return
	}

	s.Add(source)
}

func (b *RadixBlocker) Block(host string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.data == nil {
		return false
	}

	key := parser.ReverseHostName(host)

	_, ok := b.data.Get(key)
	return ok
}

func (b *RadixBlocker) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.data == nil {
		return 0
	}

	return b.data.Len()
}

func (b *RadixBlocker) Reset(source string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.data == nil {
		return
	}

	b.data.Walk(func(key string, value interface{}) bool {
		if value == nil {
			return false
		}

		if s, ok := value.(*sources); ok {
			if s.Remove(source) && s.Len() == 0 {
				b.data.Delete(key)
			}
		}

		return false
	})
}
