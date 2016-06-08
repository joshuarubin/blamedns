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
	Listen         []string             `toml:"listen"`
	Block          BlockConfig          `toml:"block"`
	ClientTimeout  Duration             `toml:"client_timeout"`
	ServerTimeout  Duration             `toml:"server_timeout"`
	DialTimeout    Duration             `toml:"dial_timeout"`
	LookupInterval Duration             `toml:"lookup_interval"`
	Cache          DNSCacheConfig       `toml:"cache"`
	Forward        []string             `toml:"forward"`
	Zone           []DNSZoneConfig      `toml:"zone"`
	Override       map[string][]string  `toml:"override"`
	OverrideTTL    Duration             `toml:"override_ttl"`
	Server         *dnsserver.DNSServer `toml:"-"`
}

func DNSFlags(c *DNSConfig) []cli.Flag {
	ret := []cli.Flag{
		altsrc.NewStringSliceFlag(cli.StringSliceFlag{
			Name:   "dns-listen",
			EnvVar: "DNS_LISTEN",
			Value:  &defaultDNSListen,
			Usage:  "url(s) to listen for dns requests on",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-client-timeout",
			EnvVar: "DNS_CLIENT_TIMEOUT",
			Value:  5 * time.Second,
			Usage:  "time to wait for remote servers to respond to queries",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-server-timeout",
			EnvVar: "DNS_SERVER_TIMEOUT",
			Value:  2 * time.Second,
			Usage:  "time for server to wait for remote clients to respond to queries",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-dial-timeout",
			EnvVar: "DNS_DIAL_TIMEOUT",
			Value:  2 * time.Second,
			Usage:  "time to wait to establish connection to remote server",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-lookup-interval",
			EnvVar: "DNS_LOOKUP_INTERVAL",
			Value:  200 * time.Millisecond,
			Usage:  "concurrency interval for lookups in miliseconds",
		}),
		altsrc.NewStringSliceFlag(cli.StringSliceFlag{
			Name:   "dns-forward",
			EnvVar: "DNS_FORWARD",
			Value:  &defaultDNSForward,
			Usage:  "default dns server(s) to forward requests to",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-override-ttl",
			EnvVar: "DNS_OVERRIDE_TTL",
			Value:  1 * time.Hour,
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

	ret = append(ret, BlockFlags()...)
	ret = append(ret, DNSCacheFlags()...)

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

func (c *DNSConfig) Parse(ctx *cli.Context) error {
	if c.Listen = ctx.StringSlice("dns-listen"); len(c.Listen) > numDefaultDNSListen {
		c.Listen = c.Listen[numDefaultDNSListen:]
	}

	if err := c.ClientTimeout.UnmarshalText([]byte(ctx.String("dns-client-timeout"))); err != nil {
		return err
	}

	if err := c.ServerTimeout.UnmarshalText([]byte(ctx.String("dns-server-timeout"))); err != nil {
		return err
	}

	if err := c.DialTimeout.UnmarshalText([]byte(ctx.String("dns-dial-timeout"))); err != nil {
		return err
	}

	if err := c.LookupInterval.UnmarshalText([]byte(ctx.String("dns-lookup-interval"))); err != nil {
		return err
	}

	if c.Forward = ctx.StringSlice("dns-forward"); len(c.Forward) > numDefaultDNSForward {
		c.Forward = c.Forward[numDefaultDNSForward:]
	}

	if err := c.OverrideTTL.UnmarshalText([]byte(ctx.String("dns-override-ttl"))); err != nil {
		return err
	}

	if err := c.Block.Parse(ctx); err != nil {
		return err
	}

	return c.Cache.Parse(ctx)
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
		Logger:         root.Logger,
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
