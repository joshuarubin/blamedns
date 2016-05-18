package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/hosts"
)

type BlockConfig struct {
	IPv4           bdtype.IP       `toml:"ipv4" cli:"ipv4,ipv4 address to return to clients for blocked a requests"`
	IPv6           bdtype.IP       `toml:"ipv6" cli:"ipv6,ipv6 address to return to clients for blocked aaaa requests"`
	Host           string          `toml:"host" cli:",host name to return to clients for blocked cname requests"`
	TTL            bdtype.Duration `toml:"ttl" cli:",ttl to return for blocked requests"`
	UpdateInterval bdtype.Duration `toml:"update_interval" cli:",update the hosts files at this interval"`
	Hosts          []string        `toml:"hosts" cli:",files in \"/etc/hosts\" format from which to derive blocked hostnames"`
	Domains        []string        `toml:"domains" cli:",files with one domain per line to block"`
	Whitelist      []string        `toml:"whitelist" cli:",domains to never block"`
	blockers       []dnsserver.Blocker
	passers        []dnsserver.Passer
}

func defaultBlockConfig() BlockConfig {
	return BlockConfig{
		Hosts: []string{
			"http://someonewhocares.org/hosts/hosts",
			"http://hosts-file.net/ad_servers.txt",
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
			"http://sysctl.org/cameleon/hosts",
		},
		Domains: []string{
			"http://mirror1.malwaredomains.com/files/justdomains",
			"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
			"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
			"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
			"https://raw.githubusercontent.com/quidsup/notrack/master/trackers.txt",
		},
		Whitelist: []string{
			"localhost",
			"localhost.localdomain",
			"broadcasthost",
			"local",
		},
	}
}

func (b BlockConfig) Block() dnsserver.Block {
	return dnsserver.Block{
		IPv4:     b.IPv4.IP(),
		IPv6:     b.IPv6.IP(),
		Host:     b.Host,
		TTL:      b.TTL.Duration(),
		Blockers: b.blockers,
		Passers:  b.passers,
	}
}

func (b *BlockConfig) Init(root *Config) error {
	b.blockers = []dnsserver.Blocker{}

	for _, t := range []struct {
		Values []string
		Parser hosts.Parser
	}{{
		Values: b.Hosts,
		Parser: hosts.FileParser,
	}, {
		Values: b.Domains,
		Parser: hosts.DomainParser,
	}} {
		for _, i := range t.Values {
			f := &hosts.Hosts{
				UpdateInterval: b.UpdateInterval.Duration(),
				Parser:         t.Parser,
				Logger:         root.Logger,
				AppName:        root.AppName,
				AppVersion:     root.AppVersion,
			}

			if err := f.Init(i, root.CacheDir); err != nil {
				return err
			}

			b.blockers = append(b.blockers, f)
		}
	}

	b.passers = []dnsserver.Passer{
		hosts.NewWhiteList(b.Whitelist),
	}

	return nil
}

func (b *BlockConfig) Start() error {
	for _, b := range b.blockers {
		if host, ok := b.(*hosts.Hosts); ok {
			host.Start()
		}
	}
	return nil
}

func (b *BlockConfig) Shutdown() error {
	for _, b := range b.blockers {
		if host, ok := b.(*hosts.Hosts); ok {
			host.Stop()
		}
	}
	return nil
}
