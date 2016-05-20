package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnscache"
	"jrubin.io/blamedns/dnsserver"
)

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
		Block: defaultBlockConfig(),
	}
}

func (cfg *DNSConfig) Init(root *Config) error {
	if err := cfg.Block.Init(root); err != nil {
		return err
	}

	cfg.Server = &dnsserver.DNSServer{
		Forward:            cfg.Forward,
		Listen:             cfg.Listen,
		Block:              cfg.Block.Block(),
		ClientTimeout:      cfg.ClientTimeout.Duration(),
		ServerTimeout:      cfg.ServerTimeout.Duration(),
		DialTimeout:        cfg.DialTimeout.Duration(),
		Interval:           cfg.Interval.Duration(),
		Logger:             root.Logger,
		CachePruneInterval: cfg.CachePruneInterval.Duration(),
		NotifyStartedFunc:  cfg.Block.Start,
		DisableDNSSEC:      cfg.DisableDNSSEC,
		DisableRecursion:   cfg.DisableRecursion,
	}

	if !cfg.DisableCache {
		cfg.Server.Cache = dnscache.NewMemory(root.Logger)
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	// cfg.Block is started by DNSServer.NotifyStartedFunc
	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() {
	cfg.Block.Shutdown()
	cfg.Server.Shutdown()
}
