package bdconfig

import (
	"fmt"
	"io"

	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/bdconfig/stringslice"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/hosts"
)

type BlockConfig struct {
	IPv4           bdtype.IP       `cli:"ipv4,ipv4 address to return to clients for blocked a requests"`
	IPv6           bdtype.IP       `cli:"ipv6,ipv6 address to return to clients for blocked aaaa requests"`
	Host           string          `cli:",host name to return to clients for blocked cname requests"`
	TTL            bdtype.Duration `cli:",ttl to return for blocked requests"`
	UpdateInterval bdtype.Duration `toml:"update_interval" cli:",update the hosts files at this interval"`
	Hosts          []string        `cli:",files in \"/etc/hosts\" format from which to derive blocked hostnames"`
	Domains        []string        `cli:",files with one domain per line to block"`
	Whitelist      []string        `cli:",domains to never block"`
	blockers       []dnsserver.Blocker
	passers        []dnsserver.Passer
}

func defaultBlockConfig() BlockConfig {
	return BlockConfig{
		Hosts: []string{
			"http://someonewhocares.org/hosts/hosts",
		},
		Domains: []string{
			"http://mirror1.malwaredomains.com/files/justdomains",
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
			"http://sysctl.org/cameleon/hosts",
			"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
			"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
			"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
			"http://hosts-file.net/ad_servers.txt",
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

func (b BlockConfig) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "[dns.block]\n")
	if err != nil {
		return n, err
	}

	var o int
	if b.IPv4.IP() != nil {
		o, err = fmt.Fprintf(w, "ipv4 = \"%s\"\n", b.IPv4)
		n += o
		if err != nil {
			return n, err
		}
	}

	if b.IPv6.IP() != nil {
		o, err = fmt.Fprintf(w, "ipv6 = \"%s\"\n", b.IPv6)
		n += o
		if err != nil {
			return n, err
		}
	}

	if len(b.Host) > 0 {
		o, err = fmt.Fprintf(w, "host = \"%s\"\n", b.Host)
		n += o
		if err != nil {
			return n, err
		}
	}

	if b.TTL > 0 {
		o, err = fmt.Fprintf(w, "ttl = \"%s\"\n", b.TTL)
		n += o
		if err != nil {
			return n, err
		}
	}

	if b.UpdateInterval > 0 {
		o, err = fmt.Fprintf(w, "update_interval = \"%s\"\n", b.UpdateInterval)
		n += o
		if err != nil {
			return n, err
		}
	}

	o, err = stringslice.Write("hosts", b.Hosts, w)
	n += o
	if err != nil {
		return n, err
	}

	o, err = stringslice.Write("domains", b.Domains, w)
	n += o
	if err != nil {
		return n, err
	}

	o, err = stringslice.Write("whitelist", b.Whitelist, w)
	n += o

	return n, err
}
