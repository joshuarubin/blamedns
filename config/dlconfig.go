package config

import (
	"time"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

var (
	defaultDLHosts = StringSlice{
		"http://someonewhocares.org/hosts/hosts",
		"http://hosts-file.net/ad_servers.txt",
		"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		// "http://sysctl.org/cameleon/hosts",
	}

	defaultDLDomains = StringSlice{
		"http://mirror1.malwaredomains.com/files/justdomains",
		"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
		"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
		"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
		"https://raw.githubusercontent.com/quidsup/notrack/master/trackers.txt",
	}
)

type DLConfig struct {
	UpdateInterval Duration    `toml:"update_interval"`
	Hosts          StringSlice `toml:"hosts"`
	Domains        StringSlice `toml:"domains"`
	DebugHTTP      bool        `toml:"debug_http"`
}

func NewDLConfig() *DLConfig {
	ret := &DLConfig{
		UpdateInterval: Duration(24 * time.Hour),
		Hosts:          make(StringSlice, len(defaultDLHosts)),
		Domains:        make(StringSlice, len(defaultDLDomains)),
	}

	copy(ret.Hosts, defaultDLHosts)
	copy(ret.Domains, defaultDLDomains)

	return ret
}

func (c *DLConfig) Flags(prefix string) []cli.Flag {
	return []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "update-interval"),
			EnvVar: envName(prefix, "UPDATE_INTERVAL"),
			Usage:  "update the downloaded files at this interval",
			Value:  &c.UpdateInterval,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "hosts"),
			EnvVar: envName(prefix, "HOSTS"),
			Value:  &c.Hosts,
			Usage:  "files to download in \"/etc/hosts\" format from which to derive blocked hostnames",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "domains"),
			EnvVar: envName(prefix, "DOMAINS"),
			Value:  &c.Domains,
			Usage:  "files to download with one domain per line to block",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        flagName(prefix, "debug-http"),
			EnvVar:      envName(prefix, "DEBUG_HTTP"),
			Usage:       "log http headers",
			Destination: &c.DebugHTTP,
		}),
	}
}
