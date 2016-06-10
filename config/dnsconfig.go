package config

import (
	"time"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var (
	defaultDNSForward = StringSlice{
		"8.8.8.8",
		"8.8.4.4",
	}

	defaultDNSListen = StringSlice{
		"udp://[::]:53",
		"tcp://[::]:53",
	}
)

type DNSConfig struct {
	Listen         StringSlice          `toml:"listen"`
	Block          *BlockConfig         `toml:"block"`
	ClientTimeout  Duration             `toml:"client_timeout"`
	ServerTimeout  Duration             `toml:"server_timeout"`
	DialTimeout    Duration             `toml:"dial_timeout"`
	LookupInterval Duration             `toml:"lookup_interval"`
	Cache          *DNSCacheConfig      `toml:"cache"`
	Forward        StringSlice          `toml:"forward"`
	Zone           DNSZones             `toml:"zone"`
	Override       StringMapStringSlice `toml:"override"`
	OverrideTTL    Duration             `toml:"override_ttl"`
}

func NewDNSConfig() *DNSConfig {
	ret := &DNSConfig{
		Listen:         make(StringSlice, len(defaultDNSListen)),
		Block:          NewBlockConfig(),
		ClientTimeout:  Duration(5 * time.Second),
		ServerTimeout:  Duration(2 * time.Second),
		DialTimeout:    Duration(2 * time.Second),
		LookupInterval: Duration(200 * time.Millisecond),
		Cache:          NewDNSCacheConfig(),
		Forward:        make(StringSlice, len(defaultDNSForward)),
		OverrideTTL:    Duration(1 * time.Hour),
	}

	copy(ret.Listen, defaultDNSListen)
	copy(ret.Forward, defaultDNSForward)

	return ret
}

func (c *DNSConfig) Flags(prefix string) []cli.Flag {
	ret := []cli.Flag{
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "listen"),
			EnvVar: envName(prefix, "LISTEN"),
			Usage:  "url(s) to listen for dns requests on",
			Value:  &c.Listen,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "client-timeout"),
			EnvVar: envName(prefix, "CLIENT_TIMEOUT"),
			Value:  &c.ClientTimeout,
			Usage:  "time to wait for remote servers to respond to queries",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "server-timeout"),
			EnvVar: envName(prefix, "SERVER_TIMEOUT"),
			Value:  &c.ServerTimeout,
			Usage:  "time for server to wait for remote clients to respond to queries",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "dial-timeout"),
			EnvVar: envName(prefix, "DIAL_TIMEOUT"),
			Value:  &c.DialTimeout,
			Usage:  "time to wait to establish connection to remote server",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "lookup-interval"),
			EnvVar: envName(prefix, "LOOKUP_INTERVAL"),
			Value:  &c.LookupInterval,
			Usage:  "concurrency interval for lookups in miliseconds",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "forward"),
			EnvVar: envName(prefix, "FORWARD"),
			Value:  &c.Forward,
			Usage:  "default dns server(s) to forward requests to",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "override-ttl"),
			EnvVar: envName(prefix, "OVERRIDE_TTL"),
			Value:  &c.OverrideTTL,
			Usage:  "ttl to return for overridden hosts",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "zone"),
			Value:  &c.Zone,
			Hidden: true,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "override"),
			Value:  &c.Override,
			Hidden: true,
		}),
	}

	ret = append(ret, c.Block.Flags(flagName(prefix, "block"))...)
	ret = append(ret, c.Cache.Flags(flagName(prefix, "cache"))...)

	return ret
}
