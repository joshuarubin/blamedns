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
	Timeout            bdtype.Duration      `toml:"timeout" cli:",time to wait for remote servers to respond to queries"`
	DialTimeout        bdtype.Duration      `toml:"dial-timeout" cli:",time to wait to establish connection to remote server"`
	Interval           bdtype.Duration      `toml:"interval" cli:",concurrency interval for lookups in miliseconds"`
	Server             *dnsserver.DNSServer `toml:"-" cli:"-"`
	CachePruneInterval bdtype.Duration      `toml:"cache_prune_interval" cli:",how often to run cache expiration"`
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
		Timeout:            cfg.Timeout.Duration(),
		DialTimeout:        cfg.DialTimeout.Duration(),
		Interval:           cfg.Interval.Duration(),
		Logger:             root.Logger,
		Cache:              dnscache.NewMemory(root.Logger),
		CachePruneInterval: cfg.CachePruneInterval.Duration(),
		NotifyStartedFunc:  cfg.Block.Start,
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() {
	cfg.Block.Shutdown()
	cfg.Server.Shutdown()
}
