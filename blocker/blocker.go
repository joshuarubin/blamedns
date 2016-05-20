package blocker

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
)

var (
	_ parser.HostAdder  = &Blocker{}
	_ dnsserver.Blocker = &Blocker{}
)

type Blocker struct {
	Logger *logrus.Logger
	data   map[string]struct{} // TODO(jrubin) this is just temporary
	mu     sync.RWMutex
}

func (b *Blocker) AddHost(source, host string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.data == nil {
		b.data = map[string]struct{}{}
	}

	b.data[host] = struct{}{}
	// TODO(jrubin)
}

func (b *Blocker) Block(host string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, found := b.data[host]
	return found
	// TODO(jrubin)
}

func (b *Blocker) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}
