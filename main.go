package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"jrubin.io/blamedns/dnsserver"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/miekg/dns"
)

const (
	defaultDNSListenAddress = "[::]:53"
	defaultConfigFile       = "/etc/blamedns/config.toml"
	defaultBlockTTL         = 1 * time.Hour
	defaultCacheDir         = "cache"
)

var (
	name, version     string
	forwardDNSServers = newStringSlice(
		"8.8.8.8",
		"8.8.4.4",
	)

	hostsFiles = newStringSlice(
		"http://someonewhocares.org/hosts/hosts",
	)

	stringSlices = []*stringSlice{
		&forwardDNSServers,
		&hostsFiles,
	}

	cfg       config
	dnsServer *dnsserver.DNSServer
	app       = cli.NewApp()
)

func init() {
	app.Name = name
	app.Version = version
	app.Usage = "" // TODO(jrubin)
	app.Authors = []cli.Author{{
		Name:  "Joshua Rubin",
		Email: "joshua@rubixconsulting.com",
	}}
	app.Before = setup
	app.Action = run
	app.Flags = append(logFlags, []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			EnvVar: "BLAMEDNS_CONFIG",
			Value:  defaultConfigFile,
			Usage:  "config file",
		},
		cli.StringFlag{
			Name:   "cache-dir",
			EnvVar: "CACHE_DIR",
			Value:  defaultCacheDir,
			Usage:  "directory in which to store cached data",
		},
		cli.StringFlag{
			Name:   "dns-udp-listen-address",
			EnvVar: "DNS_UDP_LISTEN_ADDRESS",
			Value:  defaultDNSListenAddress,
			Usage:  "ip_address:port that the dns server will listen for udp requests on",
		},
		cli.StringFlag{
			Name:   "dns-tcp-listen-address",
			EnvVar: "DNS_TCP_LISTEN_ADDRESS",
			Value:  defaultDNSListenAddress,
			Usage:  "ip_address:port that the dns server will listen for tcp requests on",
		},
		forwardDNSServers.Flag(flag{
			Name:   "forward",
			EnvVar: "FORWARD_DNS_SERVERS",
			Usage:  "dns server(s) to forward requests to",
		}),
		cli.StringFlag{
			Name:   "block-ipv4",
			EnvVar: "BLOCK_IPV4",
			Usage:  "ipv4 address to return to clients for blocked a requests",
		},
		cli.StringFlag{
			Name:   "block-ipv6",
			EnvVar: "BLOCK_IPV6",
			Usage:  "ipv6 address to return to clients for blocked aaaa requests",
		},
		cli.StringFlag{
			Name:   "block-host",
			EnvVar: "BLOCK_HOST",
			Usage:  "host name to return to clients for blocked cname requests",
		},
		cli.DurationFlag{
			Name:   "block-ttl",
			EnvVar: "BLOCK_TTL",
			Value:  defaultBlockTTL,
			Usage:  "ttl to return for blocked requests",
		},
		hostsFiles.Flag(flag{
			Name:   "hosts",
			EnvVar: "HOSTS_FILES",
			Usage:  "files in \"/etc/hosts\" format from which to derive blocked hostnames",
		}),
	}...)

	app.Commands = []cli.Command{{
		Name:   "config",
		Usage:  "write the config to stdout",
		Action: writeConfig,
	}}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(c *cli.Context) error {
	for _, ss := range stringSlices {
		if err := ss.parseFlag(c); err != nil {
			return err
		}
	}

	cd := c.String("cache-dir")
	setCfg(&cfg.CacheDir, cd, cd == defaultCacheDir)

	// logConfig
	if err := cfg.Log.parseFlags(c); err != nil {
		return err
	}

	// dnsConfig
	setCfg(&cfg.DNS.Forward, forwardDNSServers.Values, forwardDNSServers.IsDefault())

	if len(cfg.DNS.Forward) == 0 {
		return fmt.Errorf("at least one %s value required", forwardDNSServers.Name)
	}

	if cfg.DNS.Servers == nil {
		cfg.DNS.Servers = map[string]*dns.Server{}
	}

	for _, net := range []string{"udp", "tcp"} {
		if addr := c.String("dns-" + net + "-listen-address"); len(addr) > 0 {
			if _, exists := cfg.DNS.Servers[net]; !exists || addr != defaultDNSListenAddress {
				cfg.DNS.Servers[net] = &dns.Server{
					Addr: addr,
					Net:  net,
				}
			}
		}
	}

	cfg.DNS.SetDefaults()

	// if len(dnsUDPListenAddress) == 0 && len(dnsTCPListenAddress) == 0 {
	// 	return fmt.Errorf("at least one of dns-udp-listen-address or dns-tcp-listen-address is required")
	// }
	//
	// if len(dnsUDPListenAddress) == 0 {
	// 	log.Warn("no udp dns server defined")
	// }
	//
	// if len(dnsTCPListenAddress) == 0 {
	// 	log.Info("no tcp dns server defined")
	// }

	// blockConfig
	setCfg(&cfg.DNS.Block.IPv4, ip(net.ParseIP(c.String("block-ipv4"))), false)
	setCfg(&cfg.DNS.Block.IPv6, ip(net.ParseIP(c.String("block-ipv6"))), false)
	setCfg(&cfg.DNS.Block.Host, c.String("block-host"), false)

	i := c.Duration("block-ttl")
	setCfg(&cfg.DNS.Block.TTL, duration(i), i == defaultBlockTTL)

	setCfg(&cfg.DNS.Block.Hosts, hostsFiles.Values, hostsFiles.IsDefault())

	return nil
}

func isInitial(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func setCfg(cfgValue, value interface{}, isDefault bool) {
	v := reflect.ValueOf(value)

	// if value is go's "initial value, don't write it
	if isInitial(v) {
		return
	}

	cur := reflect.Indirect(reflect.ValueOf(cfgValue))

	// if the value in cur is go's "initial" value or isDefault is false it is
	// safe to overwrite
	if !isDefault || isInitial(cur) {
		cur.Set(v)
	}
}

func parseConfigFile(file string) {
	if len(file) == 0 {
		return
	}

	md, err := toml.DecodeFile(file, &cfg)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err != syscall.ENOENT {
			log.WithError(err).Warn("error reading config file")
		}
		return
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		log.Warnf("undecoded keys: %q", undecoded)
	}
}

func setup(c *cli.Context) error {
	parseConfigFile(c.String("config"))

	if err := parseFlags(c); err != nil {
		return err
	}

	cfg.Log.init()
	dnsServer = cfg.DNS.Server()

	return nil
}

func run(c *cli.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- dnsServer.ListenAndServe()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case sig := <-sigs:
		log.WithField("signal", sig).Debug("received signal")
	}

	return dnsServer.Shutdown()
}

func writeConfig(c *cli.Context) error {
	_, err := cfg.Write(os.Stdout)
	return err
}
