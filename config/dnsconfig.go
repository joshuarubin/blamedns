package config

import (
	"time"

	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/override"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var (
	defaultDNSForward = cli.StringSlice{
		"8.8.8.8",
		"8.8.4.4",
	}
	numDefaultDNSForward = len(defaultDNSForward)

	defaultDNSListen = cli.StringSlice{
		"udp://[::]:53",
		"tcp://[::]:53",
	}
	numDefaultDNSListen = len(defaultDNSListen)
)

type DNSZoneConfig struct {
	Name string   `toml:"name" json:"name"`
	Addr []string `toml:"addr" json:"addr"`
}

type DNSConfig struct {
	Listen         cli.StringSlice      `toml:"listen"`
	Block          *BlockConfig         `toml:"block"`
	ClientTimeout  Duration             `toml:"client_timeout"`
	ServerTimeout  Duration             `toml:"server_timeout"`
	DialTimeout    Duration             `toml:"dial_timeout"`
	LookupInterval Duration             `toml:"lookup_interval"`
	Cache          *DNSCacheConfig      `toml:"cache"`
	Forward        cli.StringSlice      `toml:"forward"`
	Zone           []DNSZoneConfig      `toml:"zone"`
	Override       map[string][]string  `toml:"override"`
	OverrideTTL    Duration             `toml:"override_ttl"`
	Server         *dnsserver.DNSServer `toml:"-"`
}

func NewDNSConfig() *DNSConfig {
	ret := &DNSConfig{
		Listen:         make(cli.StringSlice, len(defaultDNSListen)),
		Block:          NewBlockConfig(),
		ClientTimeout:  Duration(5 * time.Second),
		ServerTimeout:  Duration(2 * time.Second),
		DialTimeout:    Duration(2 * time.Second),
		LookupInterval: Duration(200 * time.Millisecond),
		Cache:          NewDNSCacheConfig(),
		Forward:        make(cli.StringSlice, len(defaultDNSForward)),
		OverrideTTL:    Duration(1 * time.Hour),
	}

	copy(ret.Listen, defaultDNSListen)
	copy(ret.Forward, defaultDNSForward)

	return ret
}

func (c *DNSConfig) Flags() []cli.Flag {
	ret := []cli.Flag{
		&StringSliceFlag{
			Name:        "dns-listen",
			EnvVar:      "DNS_LISTEN",
			Usage:       "url(s) to listen for dns requests on",
			Value:       c.Listen,
			Destination: &c.Listen,
		},
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-client-timeout",
			EnvVar: "DNS_CLIENT_TIMEOUT",
			Value:  &c.ClientTimeout,
			Usage:  "time to wait for remote servers to respond to queries",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-server-timeout",
			EnvVar: "DNS_SERVER_TIMEOUT",
			Value:  &c.ServerTimeout,
			Usage:  "time for server to wait for remote clients to respond to queries",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-dial-timeout",
			EnvVar: "DNS_DIAL_TIMEOUT",
			Value:  &c.DialTimeout,
			Usage:  "time to wait to establish connection to remote server",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-lookup-interval",
			EnvVar: "DNS_LOOKUP_INTERVAL",
			Value:  &c.LookupInterval,
			Usage:  "concurrency interval for lookups in miliseconds",
		}),
		&StringSliceFlag{
			Name:        "dns-forward",
			EnvVar:      "DNS_FORWARD",
			Value:       c.Forward,
			Destination: &c.Forward,
			Usage:       "default dns server(s) to forward requests to",
		},
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-override-ttl",
			EnvVar: "DNS_OVERRIDE_TTL",
			Value:  &c.OverrideTTL,
			Usage:  "ttl to return for overridden hosts",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-zone",
			Value:  JSON{&c.Zone},
			Hidden: true,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-override",
			Value:  JSON{&c.Override},
			Hidden: true,
		}),
	}

	ret = append(ret, c.Block.Flags()...)
	ret = append(ret, c.Cache.Flags()...)

	return ret
}

func (c *DNSConfig) Get(name string) (interface{}, bool) {
	switch name {
	case "dns-listen":
		return c.Listen, true
	case "dns-client-timeout":
		return c.ClientTimeout, true
	case "dns-server-timeout":
		return c.ServerTimeout, true
	case "dns-dial-timeout":
		return c.DialTimeout, true
	case "dns-lookup-interval":
		return c.LookupInterval, true
	case "dns-forward":
		return c.Forward, true
	case "dns-override-ttl":
		return c.OverrideTTL, true
	case "dns-zone":
		return c.Zone, true
	case "dns-override":
		return c.Override, true
	}

	if ret, ok := c.Block.Get(name); ok {
		return ret, ok
	}

	if ret, ok := c.Cache.Get(name); ok {
		return ret, ok
	}

	return nil, false
}

func (c *DNSConfig) Init(root *Config, onStart func()) error {
	if err := c.Block.Init(root); err != nil {
		return err
	}

	c.Cache.Init(root)

	c.Server = &dnsserver.DNSServer{
		Listen:         c.Listen,
		Block:          c.Block.Block(root),
		ClientTimeout:  c.ClientTimeout.Duration(),
		ServerTimeout:  c.ServerTimeout.Duration(),
		DialTimeout:    c.DialTimeout.Duration(),
		LookupInterval: c.LookupInterval.Duration(),
		Logger:         root.Logger.WithField("system", "dns"),
		NotifyStartedFunc: func() error {
			c.Cache.Start()
			onStart()
			return c.Block.Start()
		},
		Cache: c.Cache.Cache,
		Zones: map[string][]string{
			".": c.Forward,
		},
		OverrideTTL: c.OverrideTTL.Duration(),
		Override:    override.New(override.Parse(c.Override)),
	}

	for _, zone := range c.Zone {
		c.Server.Zones[zone.Name] = zone.Addr
	}

	return nil
}

func (c DNSConfig) Start() error {
	// c.Block and c.Cache are started by DNSServer.NotifyStartedFunc
	return c.Server.ListenAndServe()
}

func (c DNSConfig) SIGUSR1() {
	c.Cache.SIGUSR1()
}

func (c DNSConfig) Shutdown() {
	c.Cache.Shutdown()
	c.Block.Shutdown()
	c.Server.Shutdown()
}
