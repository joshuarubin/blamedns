package bdconfig

import (
	"net/url"
	"path"

	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/blocker"
	"jrubin.io/blamedns/dl"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/watcher"
	"jrubin.io/blamedns/whitelist"
)

type BlockConfig struct {
	IPv4           bdtype.IP       `toml:"ipv4" cli:"ipv4,ipv4 address to return to clients for blocked a requests"`
	IPv6           bdtype.IP       `toml:"ipv6" cli:"ipv6,ipv6 address to return to clients for blocked aaaa requests"`
	TTL            bdtype.Duration `toml:"ttl" cli:",ttl to return for blocked requests"`
	UpdateInterval bdtype.Duration `toml:"update_interval" cli:",update the hosts files at this interval"`
	Hosts          []string        `toml:"hosts" cli:",files in \"/etc/hosts\" format from which to derive blocked hostnames"`
	Domains        []string        `toml:"domains" cli:",files with one domain per line to block"`
	Whitelist      []string        `toml:"whitelist" cli:",domains to never block"`
	DebugHTTP      bool            `toml:"debug-http" cli:",log http headers"`
	dl             []*dl.DL
	blocker        *blocker.Blocker
	watchers       []*watcher.Watcher
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
		IPv4:    b.IPv4.IP(),
		IPv6:    b.IPv6.IP(),
		TTL:     b.TTL.Duration(),
		Blocker: b.blocker,
		Passer:  whitelist.New(b.Whitelist),
	}
}

func (b *BlockConfig) Init(root *Config) error {
	hostsDir := path.Join(root.CacheDir, "hosts")
	domainsDir := path.Join(root.CacheDir, "domains")

	b.blocker = &blocker.Blocker{Logger: root.Logger}

	hostsFileParser := parser.HostsFileParser{
		HostAdder: b.blocker,
		Logger:    root.Logger,
	}

	hostsWatcher, err := watcher.New(root.Logger, hostsFileParser, hostsDir)
	if err != nil {
		return err
	}

	domainParser := parser.DomainParser{
		HostAdder: b.blocker,
		Logger:    root.Logger,
	}

	domainsWatcher, err := watcher.New(root.Logger, domainParser, domainsDir)
	if err != nil {
		return err
	}

	b.watchers = []*watcher.Watcher{
		hostsWatcher,
		domainsWatcher,
	}

	for _, t := range []struct {
		Values  []string
		BaseDir string
	}{{
		Values:  b.Hosts,
		BaseDir: hostsDir,
	}, {
		Values:  b.Domains,
		BaseDir: domainsDir,
	}} {
		for _, u := range t.Values {
			p, err := url.Parse(u)
			if err != nil {
				return err
			}

			d := &dl.DL{
				URL:            p,
				BaseDir:        t.BaseDir,
				UpdateInterval: b.UpdateInterval.Duration(),
				Logger:         root.Logger,
				AppName:        root.AppName,
				AppVersion:     root.AppVersion,
				DebugHTTP:      b.DebugHTTP,
			}

			if err = d.Init(); err != nil {
				return err
			}

			b.dl = append(b.dl, d)
		}
	}

	return nil
}

func (b *BlockConfig) Start() error {
	for _, w := range b.watchers {
		w.Start()
	}

	for _, d := range b.dl {
		d.Start()
	}

	return nil
}

func (b *BlockConfig) Shutdown() {
	for _, w := range b.watchers {
		w.Stop()
	}

	for _, d := range b.dl {
		d.Stop()
	}
}
