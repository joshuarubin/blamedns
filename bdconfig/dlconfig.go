package bdconfig

import (
	"net/url"
	"path"

	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/blamedns/dl"
)

type DLConfig struct {
	UpdateInterval bdtype.Duration `toml:"update_interval" cli:",update the hosts files at this interval"`
	Hosts          []string        `toml:"hosts" cli:",files to download in \"/etc/hosts\" format from which to derive blocked hostnames"`
	Domains        []string        `toml:"domains" cli:",files to download with one domain per line to block"`
	DebugHTTP      bool            `toml:"debug_http" cli:",log http headers"`
	dl             []*dl.DL
}

var defaultDLConfig = DLConfig{
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
}

func (cfg *DLConfig) Init(root *Config) error {
	hostsDir := path.Join(root.CacheDir, "hosts")
	domainsDir := path.Join(root.CacheDir, "domains")

	for _, t := range []struct {
		Values  []string
		BaseDir string
	}{{
		Values:  cfg.Hosts,
		BaseDir: hostsDir,
	}, {
		Values:  cfg.Domains,
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
				UpdateInterval: cfg.UpdateInterval.Duration(),
				Logger:         root.Logger,
				AppName:        root.AppName,
				AppVersion:     root.AppVersion,
				DebugHTTP:      cfg.DebugHTTP,
			}

			if err = d.Init(); err != nil {
				return err
			}

			cfg.dl = append(cfg.dl, d)
		}
	}

	return nil
}

func (cfg *DLConfig) Start() {
	for _, d := range cfg.dl {
		d.Start()
	}
}

func (cfg *DLConfig) Shutdown() {
	for _, d := range cfg.dl {
		d.Stop()
	}
}
