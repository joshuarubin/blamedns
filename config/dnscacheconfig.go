package config

import (
	"time"

	"jrubin.io/blamedns/dnscache"

	"github.com/joshuarubin/cli"
	"github.com/joshuarubin/cli/altsrc"
)

type DNSCacheConfig struct {
	PruneInterval Duration         `toml:"prune_interval"`
	Disable       bool             `toml:"disable"`
	Size          int              `toml:"size"`
	Cache         *dnscache.Memory `toml:"-"`
}

func DNSCacheFlags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-cache-prune-interval",
			EnvVar: "DNS_CACHE_PRUNE_INTERVAL",
			Value:  1 * time.Hour,
			Usage:  "how often to run cache expiration",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:   "dns-cache-disable",
			EnvVar: "DNS_CACHE_DISABLE",
			Usage:  "maximum number of dns records to cache",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "dns-cache-size",
			EnvVar: "DNS_CACHE_SIZE",
			Value:  131072, // 2^17
			Usage:  "maximum number of dns records to cache",
		}),
	}
}

func (c *DNSCacheConfig) Get(name string) (interface{}, bool) {
	switch name {
	case "dns-cache-prune-interval":
		return c.PruneInterval, true
	case "dns-cache-disable":
		return c.Disable, true
	case "dns-cache-size":
		return c.Size, true
	}
	return nil, false
}

func (c *DNSCacheConfig) Parse(ctx *cli.Context) error {
	if err := c.PruneInterval.UnmarshalText([]byte(ctx.String("dns-cache-prune-interval"))); err != nil {
		return err
	}

	c.Disable = ctx.Bool("dns-cache-disable")
	c.Size = ctx.Int("dns-cache-size")

	return nil
}

func (c *DNSCacheConfig) Init(root *Config) {
	if !c.Disable {
		c.Cache = dnscache.NewMemory(c.Size, root.Logger)
	}
}

func (c DNSCacheConfig) Start() {
	if c.Cache != nil {
		c.Cache.Start(c.PruneInterval.Duration())
	}
}

func (c *DNSCacheConfig) SIGUSR1() {
	if c.Cache != nil {
		c.Cache.Purge()
	}
}

func (c DNSCacheConfig) Shutdown() {
	if c.Cache != nil {
		c.Cache.Stop()
	}
}
