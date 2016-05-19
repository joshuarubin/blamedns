package bdconfig

import (
	"net/url"

	"github.com/miekg/dns"

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

func parseDNSServer(val string) (*dns.Server, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}

	return &dns.Server{
		Addr: u.Host,
		Net:  u.Scheme,
	}, nil
}

func (cfg *DNSConfig) Init(root *Config) error {
	servers := make([]*dns.Server, len(cfg.Listen))

	for i, listen := range cfg.Listen {
		server, err := parseDNSServer(listen)
		if err != nil {
			return err
		}
		servers[i] = server
	}

	if err := cfg.Block.Init(root); err != nil {
		return err
	}

	cfg.Server = &dnsserver.DNSServer{
		Forward:            cfg.Forward,
		Servers:            servers,
		Block:              cfg.Block.Block(),
		Timeout:            cfg.Timeout.Duration(),
		DialTimeout:        cfg.DialTimeout.Duration(),
		Interval:           cfg.Interval.Duration(),
		Logger:             root.Logger,
		Cache:              dnscache.NewMemory(root.Logger),
		CachePruneInterval: cfg.CachePruneInterval.Duration(),
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	if err := cfg.Block.Start(); err != nil {
		return err
	}

	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() error {
	if err := cfg.Server.Shutdown(); err != nil {
		return err
	}

	return cfg.Block.Shutdown()
}
