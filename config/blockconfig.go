package config

import (
	"net"
	"path"
	"time"

	"jrubin.io/blamedns/blocker"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/watcher"
	"jrubin.io/blamedns/whitelist"

	"github.com/joshuarubin/cli"
	"github.com/joshuarubin/cli/altsrc"
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
	IPv4      net.IP   `toml:"ipv4"`
	IPv6      net.IP   `toml:"ipv6"`
	TTL       Duration `toml:"ttl"`
	WhiteList []string `toml:"whitelist"`
	blocker   blocker.Blocker
	watchers  []*watcher.Watcher
}

func BlockFlags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "dns-block-ipv4",
			EnvVar: "DNS_BLOCK_IPV4",
			Value:  "127.0.0.1",
			Usage:  "ipv4 address to return to clients for blocked a requests",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "dns-block-ipv6",
			EnvVar: "DNS_BLOCK_IPV6",
			Value:  "::1",
			Usage:  "ipv6 address to return to clients for blocked aaaa requests",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "dns-block-ttl",
			EnvVar: "DNS_BLOCK_TTL",
			Value:  1 * time.Hour,
			Usage:  "ttl to return for blocked requests",
		}),
		altsrc.NewStringSliceFlag(cli.StringSliceFlag{
			Name:   "dns-block-whitelist",
			EnvVar: "DNS_BLOCK_WHITELIST",
			Value:  &defaultDNSBlockWhiteList,
			Usage:  "domains to never block",
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

func (c *BlockConfig) Parse(ctx *cli.Context) error {
	c.IPv4 = net.ParseIP(ctx.String("dns-block-ipv4"))
	c.IPv6 = net.ParseIP(ctx.String("dns-block-ipv6"))

	if err := c.TTL.UnmarshalText([]byte(ctx.String("dns-block-ttl"))); err != nil {
		return err
	}

	if c.WhiteList = ctx.StringSlice("dns-block-whitelist"); len(c.WhiteList) > numDefaultDNSBlockWhiteList {
		c.WhiteList = c.WhiteList[numDefaultDNSBlockWhiteList:]
	}

	return nil
}

func (c *BlockConfig) Init(root *Config) error {
	hostsDir := path.Join(root.CacheDir, "hosts")
	domainsDir := path.Join(root.CacheDir, "domains")

	c.blocker = &blocker.RadixBlocker{}

	hostsFileParser := parser.HostsFileParser{
		HostAdder: c.blocker,
		Logger:    root.Logger,
	}

	hostsWatcher, err := watcher.New(root.Logger, hostsFileParser, hostsDir)
	if err != nil {
		return err
	}

	domainParser := parser.DomainParser{
		HostAdder: c.blocker,
		Logger:    root.Logger,
	}

	domainsWatcher, err := watcher.New(root.Logger, domainParser, domainsDir)
	if err != nil {
		return err
	}

	c.watchers = []*watcher.Watcher{
		hostsWatcher,
		domainsWatcher,
	}

	return nil
}

func (c BlockConfig) Block(root *Config) dnsserver.Block {
	return dnsserver.Block{
		IPv4:    c.IPv4,
		IPv6:    c.IPv6,
		TTL:     c.TTL.Duration(),
		Blocker: c.blocker,
		Passer:  whitelist.New(c.WhiteList...),
		Logger:  root.Logger,
	}
}

func (c *BlockConfig) Start() error {
	for _, w := range c.watchers {
		w.Start()
	}

	return nil
}

func (c *BlockConfig) Shutdown() {
	for _, w := range c.watchers {
		w.Stop()
	}
}
