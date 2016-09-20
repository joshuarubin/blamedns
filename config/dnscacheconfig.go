package config

import (
	"time"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

type DNSCacheConfig struct {
	PruneInterval Duration `toml:"prune_interval"`
	Disable       bool     `toml:"disable"`
	Size          int      `toml:"size"`
}

func NewDNSCacheConfig() *DNSCacheConfig {
	return &DNSCacheConfig{
		PruneInterval: Duration(1 * time.Hour),
		Size:          131072, // 2^17
	}
}

func (c *DNSCacheConfig) Flags(prefix string) []cli.Flag {
	return []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "prune-interval"),
			EnvVar: envName(prefix, "PRUNE_INTERVAL"),
			Value:  &c.PruneInterval,
			Usage:  "how often to run cache expiration",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        flagName(prefix, "disable"),
			EnvVar:      envName(prefix, "DISABLE"),
			Usage:       "maximum number of dns records to cache",
			Destination: &c.Disable,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:        flagName(prefix, "size"),
			EnvVar:      envName(prefix, "SIZE"),
			Usage:       "maximum number of dns records to cache",
			Value:       c.Size,
			Destination: &c.Size,
		}),
	}
}
