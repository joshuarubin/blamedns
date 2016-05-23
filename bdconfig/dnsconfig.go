package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnscache"
	"jrubin.io/blamedns/dnsserver"
)

const defaultCacheSize = 131072 // 2^17

type DNSConfig struct {
	Forward            []string             `toml:"forward" cli:",dns server(s) to forward requests to"`
	Listen             []string             `toml:"listen" cli:",url(s) to listen for dns requests on"`
	Block              BlockConfig          `toml:"block"`
	ClientTimeout      bdtype.Duration      `toml:"client_timeout" cli:",time to wait for remote servers to respond to queries"`
	ServerTimeout      bdtype.Duration      `toml:"server_timeout" cli:",time for server to wait for remote clients to respond to queries"`
	DialTimeout        bdtype.Duration      `toml:"dial_timeout" cli:",time to wait to establish connection to remote server"`
	Interval           bdtype.Duration      `toml:"interval" cli:",concurrency interval for lookups in miliseconds"`
	Server             *dnsserver.DNSServer `toml:"-" cli:"-"`
	CachePruneInterval bdtype.Duration      `toml:"cache_prune_interval" cli:",how often to run cache expiration"`
	DisableDNSSEC      bool                 `toml:"disable_dnssec" cli:",disable dnssec validation"`
	DisableRecursion   bool                 `toml:"disable_recursion" cli:",refuse recursive queries (you probably don't want to set this)"`
	DisableCache       bool                 `toml:"disable_cache" cli:",disable dns lookup cache"`
	CacheSize          int                  `toml:"cache_size" cli:",maximum number of dns records to cache"`
}

func defaultDNSConfig() DNSConfig {
	return DNSConfig{
		Forward: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
		Listen: []string{
			"udp://[::]:53",
			"tcp://[::]:53",
		},
		Block:     defaultBlockConfig(),
		CacheSize: defaultCacheSize,
	}
}

func (cfg *DNSConfig) Init(root *Config) error {
	if err := cfg.Block.Init(root); err != nil {
		return err
	}

	var cache *dnscache.Memory
	if !cfg.DisableCache {
		cache = dnscache.NewMemory(cfg.CacheSize, root.Logger, nil)
	}

	cfg.Server = &dnsserver.DNSServer{
		Forward:       cfg.Forward,
		Listen:        cfg.Listen,
		Block:         cfg.Block.Block(),
		ClientTimeout: cfg.ClientTimeout.Duration(),
		ServerTimeout: cfg.ServerTimeout.Duration(),
		DialTimeout:   cfg.DialTimeout.Duration(),
		Interval:      cfg.Interval.Duration(),
		Logger:        root.Logger,
		NotifyStartedFunc: func() error {
			if !cfg.DisableCache {
				cache.Start(cfg.CachePruneInterval.Duration())
			}
			return cfg.Block.Start()
		},
		DisableDNSSEC:    cfg.DisableDNSSEC,
		DisableRecursion: cfg.DisableRecursion,
		Cache:            cache,
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	// cfg.Block is started by DNSServer.NotifyStartedFunc
	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() {
	if !cfg.DisableCache {
		if cache, ok := cfg.Server.Cache.(*dnscache.Memory); ok && cache != nil {
			cache.Stop()
		}
	}

	cfg.Block.Shutdown()
	cfg.Server.Shutdown()
}
