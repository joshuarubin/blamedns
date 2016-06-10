package config

import (
	"time"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var (
	defaultDNSBlockWhiteList = cli.StringSlice{
		"localhost",
		"localhost.localdomain",
		"broadcasthost",
		"local",
	}
	numDefaultDNSBlockWhiteList = len(defaultDNSBlockWhiteList)
)

type BlockConfig struct {
	IPv4      IP          `toml:"ipv4"`
	IPv6      IP          `toml:"ipv6"`
	TTL       Duration    `toml:"ttl"`
	WhiteList StringSlice `toml:"whitelist"`
}

func NewBlockConfig() *BlockConfig {
	ret := &BlockConfig{
		IPv4:      ParseIP("127.0.0.1"),
		IPv6:      ParseIP("::1"),
		TTL:       Duration(1 * time.Hour),
		WhiteList: make(StringSlice, len(defaultDNSBlockWhiteList)),
	}

	copy(ret.WhiteList, defaultDNSBlockWhiteList)

	return ret
}

func (c *BlockConfig) Flags(prefix string) []cli.Flag {
	return []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "ipv4"),
			EnvVar: envName(prefix, "IPV4"),
			Usage:  "ipv4 address to return to clients for blocked a requests",
			Value:  &c.IPv4,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "ipv6"),
			EnvVar: envName(prefix, "IPV6"),
			Usage:  "ipv6 address to return to clients for blocked aaaa requests",
			Value:  &c.IPv6,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "ttl"),
			EnvVar: envName(prefix, "TTL"),
			Usage:  "ttl to return for blocked requests",
			Value:  &c.TTL,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "whitelist"),
			EnvVar: envName(prefix, "WHITELIST"),
			Usage:  "domains to never block",
			Value:  &c.WhiteList,
		}),
	}
}
