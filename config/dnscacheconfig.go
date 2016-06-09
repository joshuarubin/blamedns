package config

import (
	"time"

	"jrubin.io/blamedns/dnscache"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

type DNSCacheConfig struct {
	PruneInterval Duration         `toml:"prune_interval"`
	Disable       bool             `toml:"disable"`
	Size          int              `toml:"size"`
	Cache         *dnscache.Memory `toml:"-"`
}

func NewDNSCacheConfig() *DNSCacheConfig {
	return &DNSCacheConfig{
		PruneInterval: Duration(1 * time.Hour),
		Size:          131072, // 2^17
	}
}

func (c *DNSCacheConfig) Flags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-cache-prune-interval",
			EnvVar: "DNS_CACHE_PRUNE_INTERVAL",
			Value:  &c.PruneInterval,
			Usage:  "how often to run cache expiration",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        "dns-cache-disable",
			EnvVar:      "DNS_CACHE_DISABLE",
			Usage:       "maximum number of dns records to cache",
			Destination: &c.Disable,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:        "dns-cache-size",
			EnvVar:      "DNS_CACHE_SIZE",
			Usage:       "maximum number of dns records to cache",
			Value:       c.Size,
			Destination: &c.Size,
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
