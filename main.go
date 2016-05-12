package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jrubin.io/blamedns/dnsserver"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/miekg/dns"
)

const (
	name                    = "blamedns"
	version                 = "0.1.0"
	defaultDNSListenAddress = "[::]:53"
)

var (
	defaultForwardDNSServers = cli.StringSlice{
		"8.8.8.8",
		"8.8.4.4",
	}
	forwardDNSServers cli.StringSlice
	dnsServer         *dnsserver.DNSServer
	app               = cli.NewApp()
)

func init() {
	forwardDNSServers = make(cli.StringSlice, len(defaultForwardDNSServers))
	copy(forwardDNSServers, defaultForwardDNSServers)

	app.Name = name
	app.Version = version
	app.Usage = "" // TODO(jrubin)
	app.Authors = []cli.Author{{
		Name:  "Joshua Rubin",
		Email: "joshua@rubixconsulting.com",
	}}
	app.Before = setup
	app.Action = run
	app.Flags = []cli.Flag{
		logFlag,
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
		cli.StringSliceFlag{
			Name:   "forward",
			EnvVar: "FORWARD_DNS_SERVERS",
			Value:  &forwardDNSServers,
			Usage:  "dns server(s) to forward requests to",
		},
		cli.StringFlag{
			Name:   "block-ip",
			EnvVar: "BLOCK_IP",
			Usage:  "ip address to return to clients for blocked requests",
		},
		cli.DurationFlag{
			Name:   "block-ttl",
			EnvVar: "BLOCK_TTL",
			Value:  3600 * time.Second,
			Usage:  "ttl to return for blocked requests",
		},
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func stripEmpty(s []string) []string {
	var ret []string

	for _, v := range s {
		if len(v) > 0 {
			ret = append(ret, v)
		}
	}

	return ret
}

func setup(c *cli.Context) error {
	setLogLevel(c)

	forwardDNSServers = c.StringSlice("forward")

	// strip out the default values if the user supplied any
	if len(forwardDNSServers) > len(defaultForwardDNSServers) {
		forwardDNSServers = forwardDNSServers[len(defaultForwardDNSServers):]
	}

	dnsUDPListenAddress := c.String("dns-udp-listen-address")
	dnsTCPListenAddress := c.String("dns-tcp-listen-address")

	if len(dnsUDPListenAddress) == 0 && len(dnsTCPListenAddress) == 0 {
		return fmt.Errorf("at least one of dns-udp-listen-address or dns-tcp-listen-address is required")
	}

	if len(dnsUDPListenAddress) == 0 {
		log.Warn("no udp dns server defined")
	}

	if len(dnsTCPListenAddress) == 0 {
		log.Info("no tcp dns server defined")
	}

	forwardDNSServers = stripEmpty(forwardDNSServers)
	if len(forwardDNSServers) == 0 {
		return fmt.Errorf("at least one forward dns server is required")
	}

	dnsServer = &dnsserver.DNSServer{
		Servers: []*dns.Server{
			{Addr: dnsUDPListenAddress, Net: "udp"},
			{Addr: dnsTCPListenAddress, Net: "tcp"},
		},
		Forward: forwardDNSServers,
		Block: dnsserver.Block{
			IP:  net.ParseIP(c.String("block-ip")),
			TTL: c.Duration("block-ttl"),
		},
	}

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
