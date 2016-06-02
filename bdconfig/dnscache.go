package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnscache"
)

const defaultDNSCacheSize = 131072 // 2^17

type DNSCacheConfig struct {
	PruneInterval bdtype.Duration  `toml:"prune_interval" cli:",how often to run cache expiration"`
	Disable       bool             `toml:"disable" cli:",disable dns lookup cache"`
	Size          int              `toml:"size" cli:",maximum number of dns records to cache"`
	Cache         *dnscache.Memory `toml:"-" cli:"-"`
}

var defaultDNSCacheConfig = DNSCacheConfig{
	Size: defaultDNSCacheSize,
}

func (cfg *DNSCacheConfig) Init(root *Config) {
	if !cfg.Disable {
		cfg.Cache = dnscache.NewMemory(cfg.Size, root.Logger)
	}
}

func (cfg DNSCacheConfig) Start() {
	if !cfg.Disable {
		cfg.Cache.Start(cfg.PruneInterval.Duration())
	}
}

func (cfg DNSCacheConfig) Shutdown() {
	if !cfg.Disable {
		cfg.Cache.Stop()
	}
}

func (cfg *DNSCacheConfig) SIGUSR1() {
	cfg.Cache.Purge()
}
