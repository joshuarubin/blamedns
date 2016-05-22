package blocker

import (
	"sort"
	"sync"
)

type hostData struct {
	Host    string
	Sources *sources
}

type SliceBlocker struct {
	mu    sync.RWMutex
	hosts []*hostData
}

func (b *SliceBlocker) search(host string) int {
	return sort.Search(len(b.hosts), func(i int) bool {
		return b.hosts[i].Host == host
	})
}

func (b *SliceBlocker) has(host string) (int, bool) {
	n := b.search(host)

	if n < len(b.hosts) && b.hosts[n].Host == host {
		return n, true
	}

	return n, false
}

func (b *SliceBlocker) AddHost(source, host string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n, exist := b.has(host)

	if exist {
		b.hosts[n].Sources.Add(source)
		return
	}

	// insert host into slice at index n
	b.hosts = append(b.hosts, nil)
	copy((b.hosts)[n+1:], (b.hosts)[n:])
	(b.hosts)[n] = &hostData{
		Host:    host,
		Sources: newSources(source),
	}
}

func (b *SliceBlocker) Block(host string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exist := b.has(host)
	return exist
}

func (b *SliceBlocker) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.hosts)
}

func (b *SliceBlocker) Reset(source string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var deleted int
	for i, hs := range b.hosts {
		if !hs.Sources.Remove(source) {
			continue
		}

		if hs.Sources.Len() == 0 {
			j := i - deleted
			b.hosts = b.hosts[:j+copy(b.hosts[j:], b.hosts[j+1:])]
			deleted++
		}
	}
}
