package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnsserver"
)

type DNSConfig struct {
	Forward         []string             `toml:"forward" cli:",dns server(s) to forward requests to"`
	Listen          []string             `toml:"listen" cli:",url(s) to listen for dns requests on"`
	Block           BlockConfig          `toml:"block"`
	ClientTimeout   bdtype.Duration      `toml:"client_timeout" cli:",time to wait for remote servers to respond to queries"`
	ServerTimeout   bdtype.Duration      `toml:"server_timeout" cli:",time for server to wait for remote clients to respond to queries"`
	DialTimeout     bdtype.Duration      `toml:"dial_timeout" cli:",time to wait to establish connection to remote server"`
	ForwardInterval bdtype.Duration      `toml:"interval" cli:",concurrency interval for lookups in miliseconds"`
	Server          *dnsserver.DNSServer `toml:"-" cli:"-"`
	DisableDNSSEC   bool                 `toml:"disable_dnssec" cli:",disable dnssec validation"`
	Cache           DNSCacheConfig       `toml:"cache"`
}

var defaultDNSConfig = DNSConfig{
	Forward: []string{
		"8.8.8.8",
		"8.8.4.4",
	},
	Listen: []string{
		"udp://[::]:53",
		"tcp://[::]:53",
	},
	Block: defaultBlockConfig,
	Cache: defaultDNSCacheConfig,
}

func (cfg *DNSConfig) Init(root *Config, onStart func()) error {
	if err := cfg.Block.Init(root); err != nil {
		return err
	}

	cfg.Cache.Init(root)

	cfg.Server = &dnsserver.DNSServer{
		Forward:         cfg.Forward,
		Listen:          cfg.Listen,
		Block:           cfg.Block.Block(),
		ClientTimeout:   cfg.ClientTimeout.Duration(),
		ServerTimeout:   cfg.ServerTimeout.Duration(),
		DialTimeout:     cfg.DialTimeout.Duration(),
		ForwardInterval: cfg.ForwardInterval.Duration(),
		Logger:          root.Logger,
		NotifyStartedFunc: func() error {
			cfg.Cache.Start()
			onStart()
			return cfg.Block.Start()
		},
		DisableDNSSEC: cfg.DisableDNSSEC,
		Cache:         cfg.Cache.Cache,
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	// cfg.Block  and cfg.Cache are started by DNSServer.NotifyStartedFunc
	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() {
	cfg.Cache.Shutdown()
	cfg.Block.Shutdown()
	cfg.Server.Shutdown()
}
