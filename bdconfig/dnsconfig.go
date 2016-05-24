package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnsserver"
)

type DNSZoneConfig struct {
	Name string   `toml:"name"`
	Addr []string `toml:"addr"`
}

type DNSConfig struct {
	Listen          []string             `toml:"listen" cli:",url(s) to listen for dns requests on"`
	Block           BlockConfig          `toml:"block"`
	ClientTimeout   bdtype.Duration      `toml:"client_timeout" cli:",time to wait for remote servers to respond to queries"`
	ServerTimeout   bdtype.Duration      `toml:"server_timeout" cli:",time for server to wait for remote clients to respond to queries"`
	DialTimeout     bdtype.Duration      `toml:"dial_timeout" cli:",time to wait to establish connection to remote server"`
	ForwardInterval bdtype.Duration      `toml:"interval" cli:",concurrency interval for lookups in miliseconds"`
	Server          *dnsserver.DNSServer `toml:"-" cli:"-"`
	DisableDNSSEC   bool                 `toml:"disable_dnssec" cli:",disable dnssec validation"`
	Cache           DNSCacheConfig       `toml:"cache"`
	StubZone        []DNSZoneConfig      `toml:"stub_zone" cli:"-"`
	ForwardZone     []DNSZoneConfig      `toml:"forward_zone" cli:"-"`
}

var defaultDNSConfig = DNSConfig{
	Listen: []string{
		"udp://[::]:53",
		"tcp://[::]:53",
	},
	Block: defaultBlockConfig,
	Cache: defaultDNSCacheConfig,
}

var defaultDNSForwardZones = []DNSZoneConfig{{
	Name: ".",
	Addr: []string{"8.8.8.8", "8.8.4.4"},
}}

func (cfg *DNSConfig) Init(root *Config, onStart func()) error {
	if err := cfg.Block.Init(root); err != nil {
		return err
	}

	cfg.Cache.Init(root)

	cfg.Server = &dnsserver.DNSServer{
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
		StubZones:     map[string][]string{},
		ForwardZones:  map[string][]string{},
	}

	for _, zone := range cfg.StubZone {
		cfg.Server.StubZones[zone.Name] = zone.Addr
	}

	for _, zone := range cfg.ForwardZone {
		cfg.Server.ForwardZones[zone.Name] = zone.Addr
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
