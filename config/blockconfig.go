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

func (c *BlockConfig) Flags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-block-ipv4",
			EnvVar: "DNS_BLOCK_IPV4",
			Usage:  "ipv4 address to return to clients for blocked a requests",
			Value:  &c.IPv4,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-block-ipv6",
			EnvVar: "DNS_BLOCK_IPV6",
			Usage:  "ipv6 address to return to clients for blocked aaaa requests",
			Value:  &c.IPv6,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-block-ttl",
			EnvVar: "DNS_BLOCK_TTL",
			Usage:  "ttl to return for blocked requests",
			Value:  &c.TTL,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dns-block-whitelist",
			EnvVar: "DNS_BLOCK_WHITELIST",
			Usage:  "domains to never block",
			Value:  &c.WhiteList,
		}),
	}
}

func (c *BlockConfig) Get(name string) (interface{}, bool) {
	switch name {
	case "dns-block-ipv4":
		return c.IPv4, true
	case "dns-block-ipv6":
		return c.IPv6, true
	case "dns-block-ttl":
		return c.TTL, true
	case "dns-block-whitelist":
		return c.WhiteList, true
	}
	return nil, false
}
