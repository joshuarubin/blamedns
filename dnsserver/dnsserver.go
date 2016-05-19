package dnsserver

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"jrubin.io/blamedns/dnscache"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

type DNSServer struct {
	Forward            []string
	Listen             []string
	servers            []*dns.Server
	Block              Block
	Logger             *logrus.Logger
	Timeout            time.Duration
	DialTimeout        time.Duration
	Interval           time.Duration
	Cache              dnscache.Cache
	CachePruneInterval time.Duration
	cachePruneTicker   *time.Ticker
	DisableDNSSEC      bool // TODO(jrubin)
	DisableRecursion   bool // TODO(jrubin)
}

// TODO(jrubin) add zone parsing and authoritative responses for them

const DefaultPort = 53

func addDefaultPort(addrs []string) ([]string, error) {
	ret := make([]string, len(addrs))

	for i, addr := range addrs {
		_, _, err := net.SplitHostPort(addr)
		if err == nil {
			ret[i] = addr
			continue
		}

		if aerr, ok := err.(*net.AddrError); ok {
			if aerr.Err == "missing port in address" {
				ret[i] = fmt.Sprintf("%s:%d", addr, DefaultPort)
				continue
			}
		}

		return nil, err
	}

	return ret, nil
}

func (d *DNSServer) startCachePruner() {
	if d.Cache == nil {
		d.Logger.Info("dns cache is disabled")
		return
	}

	d.Logger.Infof("will run cache pruning every %v", d.CachePruneInterval)
	d.cachePruneTicker = time.NewTicker(d.CachePruneInterval)

	go func() {
		for range d.cachePruneTicker.C {
			n := d.Cache.Prune()
			d.Logger.Debugf("pruned %d items from cache", n)
		}
	}()
}

func (d *DNSServer) parseDNSServer(val string) (*dns.Server, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}

	return &dns.Server{
		Addr:    u.Host,
		Net:     u.Scheme,
		Handler: d.Handler(u.Scheme),
		// TODO(jrubin)
		// ReadTimeout:  2*time.Second,
		// WriteTimeout: 2*time.Second,
	}, nil
}

func (d *DNSServer) createServers() error {
	d.servers = make([]*dns.Server, len(d.Listen))

	for i, listen := range d.Listen {
		server, err := d.parseDNSServer(listen)
		if err != nil {
			return err
		}
		d.servers[i] = server
	}
	return nil
}

func (d *DNSServer) ListenAndServe() error {
	var err error
	if err = d.createServers(); err != nil {
		return err
	}

	if d.Forward, err = addDefaultPort(d.Forward); err != nil {
		return err
	}

	d.Logger.Infof("using forwarders: %s", strings.Join(d.Forward, ", "))

	errCh := make(chan error, len(d.servers))

	for _, server := range d.servers {
		go func(server *dns.Server) {
			d.Logger.WithFields(logrus.Fields{
				"net":  server.Net,
				"addr": server.Addr,
			}).Info("starting dns server")
			errCh <- server.ListenAndServe()
		}(server)
	}

	d.startCachePruner()

	for range d.servers {
		if err := <-errCh; err != nil {
			return err
		}
	}

	return nil
}

func (d *DNSServer) stopCachePruner() {
	if d.cachePruneTicker == nil {
		return
	}

	d.cachePruneTicker.Stop()
	d.cachePruneTicker = nil
	d.Logger.Debug("stopped cache pruner")
}

func (d *DNSServer) Shutdown() error {
	d.stopCachePruner()

	for _, server := range d.servers {
		if err := server.Shutdown(); err != nil {
			return err
		}
		d.Logger.WithFields(logrus.Fields{
			"net":  server.Net,
			"addr": server.Addr,
		}).Debug("successfully shut down dns server")
	}

	return nil
}
