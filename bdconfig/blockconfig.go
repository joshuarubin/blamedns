package bdconfig

import (
	"path"

	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/blocker"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/watcher"
	"jrubin.io/blamedns/whitelist"
)

type BlockConfig struct {
	IPv4      bdtype.IP       `toml:"ipv4" cli:"ipv4,ipv4 address to return to clients for blocked a requests"`
	IPv6      bdtype.IP       `toml:"ipv6" cli:"ipv6,ipv6 address to return to clients for blocked aaaa requests"`
	TTL       bdtype.Duration `toml:"ttl" cli:",ttl to return for blocked requests"`
	Whitelist []string        `toml:"whitelist" cli:",domains to never block"`
	blocker   blocker.Blocker
	watchers  []*watcher.Watcher
}

var defaultBlockConfig = BlockConfig{
	Whitelist: []string{
		"localhost",
		"localhost.localdomain",
		"broadcasthost",
		"local",
	},
}

func (b BlockConfig) Block(root *Config) dnsserver.Block {
	return dnsserver.Block{
		IPv4:    b.IPv4.IP(),
		IPv6:    b.IPv6.IP(),
		TTL:     b.TTL.Duration(),
		Blocker: b.blocker,
		Passer:  whitelist.New(b.Whitelist),
		Logger:  root.Logger,
	}
}

func (b *BlockConfig) Init(root *Config) error {
	hostsDir := path.Join(root.CacheDir, "hosts")
	domainsDir := path.Join(root.CacheDir, "domains")

	b.blocker = &blocker.RadixBlocker{}

	hostsFileParser := parser.HostsFileParser{
		HostAdder: b.blocker,
		Logger:    root.Logger,
	}

	hostsWatcher, err := watcher.New(root.Logger, hostsFileParser, hostsDir)
	if err != nil {
		return err
	}

	domainParser := parser.DomainParser{
		HostAdder: b.blocker,
		Logger:    root.Logger,
	}

	domainsWatcher, err := watcher.New(root.Logger, domainParser, domainsDir)
	if err != nil {
		return err
	}

	b.watchers = []*watcher.Watcher{
		hostsWatcher,
		domainsWatcher,
	}

	return nil
}

func (b *BlockConfig) Start() error {
	for _, w := range b.watchers {
		w.Start()
	}

	return nil
}

func (b *BlockConfig) Shutdown() {
	for _, w := range b.watchers {
		w.Stop()
	}
}
