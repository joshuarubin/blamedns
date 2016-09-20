package config

import (
	"time"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

var (
	defaultDNSForward = StringSlice{
		"https://dns.google.com/resolve",
	}

	defaultDNSListen = StringSlice{
		"udp://[::]:53",
		"tcp://[::]:53",
	}
)

type DNSHTTPConfig struct {
	KeepAlive             Duration `toml:"keep_alive"`
	MaxIdleConns          int      `toml:"max_idle_conns"`
	MaxIdleConnsPerHost   int      `toml:"max_idle_conns_per_host"`
	IdleConnTimeout       Duration `toml:"idle_conn_timeout"`
	TLSHandshakeTimeout   Duration `toml:"tls_handshake_timeout"`
	ExpectContinueTimeout Duration `toml:"expect_continue_timeout"`
	NoDNSSEC              bool     `toml:"no_dnssec"`
	EDNSClientSubnet      string   `toml:"edns_client_subnet"`
	NoRandomPadding       bool     `toml:"no_random_padding"`
}

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
	HTTP           DNSHTTPConfig
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
		HTTP: DNSHTTPConfig{
			KeepAlive:             Duration(24 * time.Hour),
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			TLSHandshakeTimeout:   Duration(10 * time.Second),
			ExpectContinueTimeout: Duration(1 * time.Second),
		},
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
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "http-keep-alive"),
			EnvVar: envName(prefix, "HTTP_KEEP_ALIVE"),
			Value:  &c.HTTP.KeepAlive,
			Usage:  "keep idle connections alive at least this long",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:        flagName(prefix, "http-max-idle-conns"),
			EnvVar:      envName(prefix, "HTTP_MAX_IDLE_CONNS"),
			Usage:       "maximum number of idle (keep-alive) connections across all hosts. zero means no limit",
			Value:       c.HTTP.MaxIdleConns,
			Destination: &c.HTTP.MaxIdleConns,
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:        flagName(prefix, "http-max-idle-conns-per-host"),
			EnvVar:      envName(prefix, "HTTP_MAX_IDLE_CONNS_PER_HOST"),
			Usage:       "maximum number of idle (keep-alive) connections per-host. if zero, http.DefaultMaxIdleConnsPerHost is used.",
			Value:       c.HTTP.MaxIdleConnsPerHost,
			Destination: &c.HTTP.MaxIdleConnsPerHost,
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "http-idle-conn-timeout"),
			EnvVar: envName(prefix, "HTTP_IDLE_CONN_TIMEOUT"),
			Value:  &c.HTTP.IdleConnTimeout,
			Usage:  "maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. zero means no limit",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "http-tls-handshake-timeout"),
			EnvVar: envName(prefix, "HTTP_TLS_HANDSHAKE_TIMEOUT"),
			Value:  &c.HTTP.TLSHandshakeTimeout,
			Usage:  "maximum amount of time to wait for a tls handshake. zero means no timeout",
		}),
		altsrc.NewGenericFlag(cli.GenericFlag{
			Name:   flagName(prefix, "http-expect-continue-timeout"),
			EnvVar: envName(prefix, "HTTP_EXPECT_CONTINUETIMEOUT"),
			Value:  &c.HTTP.ExpectContinueTimeout,
			Usage:  "the amount of time to wait for a server's first response headers after fully writing the request headers if the request has an \"expect: 100-continue\" header. zero means no timeout",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        flagName(prefix, "http-no-dnssec"),
			EnvVar:      envName(prefix, "HTTP_NO_DNSSEC"),
			Usage:       "disable dnssec validation",
			Destination: &c.HTTP.NoDNSSEC,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        flagName(prefix, "http-edns-client-subnet"),
			EnvVar:      envName(prefix, "HTTP_EDNS_CLIENT_SUBNET"),
			Usage:       "if you are using dns-over-https because of privacy concerns, and do not want any part of your ip address to be sent to authoritative name servers for geographic location accuracy, use 0.0.0.0/0",
			Value:       c.HTTP.EDNSClientSubnet,
			Destination: &c.HTTP.EDNSClientSubnet,
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        flagName(prefix, "http-no-random-padding"),
			EnvVar:      envName(prefix, "HTTP_NO_RANDOM_PADDING"),
			Usage:       "disable random padding. random padding is used to prevent possible side-channel privacy attacks using the packet sizes of https get requests by making all requests exactly the same size",
			Destination: &c.HTTP.NoRandomPadding,
		}),
	}

	ret = append(ret, c.Block.Flags(flagName(prefix, "block"))...)
	ret = append(ret, c.Cache.Flags(flagName(prefix, "cache"))...)

	return ret
}
