package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/override"
)

type DNSZoneConfig struct {
	Name string   `toml:"name"`
	Addr []string `toml:"addr"`
}

type DNSConfig struct {
	Listen         []string             `toml:"listen" cli:",url(s) to listen for dns requests on"`
	Block          BlockConfig          `toml:"block"`
	ClientTimeout  bdtype.Duration      `toml:"client_timeout" cli:",time to wait for remote servers to respond to queries"`
	ServerTimeout  bdtype.Duration      `toml:"server_timeout" cli:",time for server to wait for remote clients to respond to queries"`
	DialTimeout    bdtype.Duration      `toml:"dial_timeout" cli:",time to wait to establish connection to remote server"`
	LookupInterval bdtype.Duration      `toml:"interval" cli:",concurrency interval for lookups in miliseconds"`
	Server         *dnsserver.DNSServer `toml:"-" cli:"-"`
	Cache          DNSCacheConfig       `toml:"cache"`
	Forward        []string             `toml:"forward" cli:",default dns server(s) to forward requests to"`
	Zone           []DNSZoneConfig      `toml:"zone" cli:"-"`
	Override       map[string][]string  `toml:"override" cli:"-"` // this should be map[string][]bdtype.IP, but that doesn't work with the toml decoder :(
	OverrideTTL    bdtype.Duration      `toml:"override_ttl" cli:",ttl to return for overridden hosts"`
}

var defaultDNSConfig = DNSConfig{
	Listen: []string{
		"udp://[::]:53",
		"tcp://[::]:53",
	},
	Block:   defaultBlockConfig,
	Cache:   defaultDNSCacheConfig,
	Forward: []string{"8.8.8.8", "8.8.4.4"},
}

func (cfg *DNSConfig) Init(root *Config, onStart func()) error {
	if err := cfg.Block.Init(root); err != nil {
		return err
	}

	cfg.Cache.Init(root)

	cfg.Server = &dnsserver.DNSServer{
		Listen:         cfg.Listen,
		Block:          cfg.Block.Block(root),
		ClientTimeout:  cfg.ClientTimeout.Duration(),
		ServerTimeout:  cfg.ServerTimeout.Duration(),
		DialTimeout:    cfg.DialTimeout.Duration(),
		LookupInterval: cfg.LookupInterval.Duration(),
		Logger:         root.Logger,
		NotifyStartedFunc: func() error {
			cfg.Cache.Start()
			onStart()
			return cfg.Block.Start()
		},
		Cache: cfg.Cache.Cache,
		Zones: map[string][]string{
			".": cfg.Forward,
		},
		OverrideTTL: cfg.OverrideTTL.Duration(),
		Override:    override.New(override.Parse(cfg.Override)),
	}

	for _, zone := range cfg.Zone {
		cfg.Server.Zones[zone.Name] = zone.Addr
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	// cfg.Block and cfg.Cache are started by DNSServer.NotifyStartedFunc
	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() {
	cfg.Cache.Shutdown()
	cfg.Block.Shutdown()
	cfg.Server.Shutdown()
}

func (cfg DNSConfig) SIGUSR1() {
	cfg.Cache.SIGUSR1()
}
