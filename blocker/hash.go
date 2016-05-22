package blocker

import "sync"

type HashBlocker struct {
	m  map[string]*sources
	mu sync.RWMutex
}

func (b *HashBlocker) AddHost(source, host string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.m == nil {
		b.m = map[string]*sources{host: newSources(source)}
		return
	}

	s, ok := b.m[host]
	if !ok || s == nil {
		b.m[host] = newSources(source)
		return
	}

	s.Add(source)
}

func (b *HashBlocker) Block(host string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.m == nil {
		return false
	}

	_, ok := b.m[host]
	return ok
}

func (b *HashBlocker) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.m)
}

func (b *HashBlocker) Reset(source string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for host, sources := range b.m {
		if !sources.Remove(source) {
			continue
		}

		if sources.Len() == 0 {
			delete(b.m, host)
		}
	}
}
