package config

import (
	"path"
	"time"

	"jrubin.io/blamedns/blocker"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/watcher"
	"jrubin.io/blamedns/whitelist"

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
	IPv4      IP              `toml:"ipv4"`
	IPv6      IP              `toml:"ipv6"`
	TTL       Duration        `toml:"ttl"`
	WhiteList cli.StringSlice `toml:"whitelist"`
	blocker   blocker.Blocker
	watchers  []*watcher.Watcher
}

func NewBlockConfig() *BlockConfig {
	ret := &BlockConfig{
		IPv4:      ParseIP("127.0.0.1"),
		IPv6:      ParseIP("::1"),
		TTL:       Duration(1 * time.Hour),
		WhiteList: make(cli.StringSlice, len(defaultDNSBlockWhiteList)),
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
		&StringSliceFlag{
			Name:        "dns-block-whitelist",
			EnvVar:      "DNS_BLOCK_WHITELIST",
			Usage:       "domains to never block",
			Value:       c.WhiteList,
			Destination: &c.WhiteList,
		},
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
		IPv4:    c.IPv4.IP(),
		IPv6:    c.IPv6.IP(),
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
