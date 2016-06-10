package config

import (
	"time"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var (
	defaultDLHosts = cli.StringSlice{
		"http://someonewhocares.org/hosts/hosts",
		"http://hosts-file.net/ad_servers.txt",
		"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		"http://sysctl.org/cameleon/hosts",
	}
	numDefaultDLHosts = len(defaultDLHosts)

	defaultDLDomains = cli.StringSlice{
		"http://mirror1.malwaredomains.com/files/justdomains",
		"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
		"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
		"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
		"https://raw.githubusercontent.com/quidsup/notrack/master/trackers.txt",
	}
	numDefaultDLDomains = len(defaultDLDomains)
)

type DLConfig struct {
	UpdateInterval Duration        `toml:"update_interval"`
	Hosts          cli.StringSlice `toml:"hosts"`
	Domains        cli.StringSlice `toml:"domains"`
	DebugHTTP      bool            `toml:"debug_http"`
}

func NewDLConfig() *DLConfig {
	ret := &DLConfig{
		UpdateInterval: Duration(24 * time.Hour),
		Hosts:          make(cli.StringSlice, len(defaultDLHosts)),
		Domains:        make(cli.StringSlice, len(defaultDLDomains)),
	}

	copy(ret.Hosts, defaultDLHosts)
	copy(ret.Domains, defaultDLDomains)

	return ret
}

func (c *DLConfig) Flags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   "dl-update-interval",
			EnvVar: "DL_UPDATE_INTERVAL",
			Usage:  "update the downloaded files at this interval",
			Value:  &c.UpdateInterval,
		}),
		&StringSliceFlag{
			Name:        "dl-hosts",
			EnvVar:      "DL_HOSTS",
			Value:       c.Hosts,
			Destination: &c.Hosts,
			Usage:       "files to download in \"/etc/hosts\" format from which to derive blocked hostnames",
		},
		&StringSliceFlag{
			Name:        "dl-domains",
			EnvVar:      "DL_DOMAINS",
			Value:       c.Domains,
			Destination: &c.Domains,
			Usage:       "files to download with one domain per line to block",
		},
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        "dl-debug-http",
			EnvVar:      "DL_DEBUG_HTTP",
			Usage:       "log http headers",
			Destination: &c.DebugHTTP,
		}),
	}
}

func (c *DLConfig) Get(name string) (interface{}, bool) {
	switch name {
	case "dl-update-interval":
		return c.UpdateInterval, true
	case "dl-hosts":
		return c.Hosts, true
	case "dl-domains":
		return c.Domains, true
	case "dl-debug-http":
		return c.DebugHTTP, true
	}

	return nil, false
}
