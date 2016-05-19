package dnsserver

import (
	"fmt"
	"net"
	"strings"
	"time"

	"jrubin.io/blamedns/dnscache"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

type DNSServer struct {
	Forward            []string
	Servers            []*dns.Server
	Block              Block
	Logger             *logrus.Logger
	Timeout            time.Duration
	Interval           time.Duration
	Cache              dnscache.Cache
	CachePruneInterval time.Duration
	cachePruneTicker   *time.Ticker
}

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

func (d *DNSServer) ListenAndServe() error {
	var err error
	if d.Forward, err = addDefaultPort(d.Forward); err != nil {
		return err
	}

	d.Logger.Infof("using forwarders: %s", strings.Join(d.Forward, ", "))

	errCh := make(chan error, len(d.Servers))

	for _, server := range d.Servers {
		server.Handler = d.Handler(server.Net)
		go func(server *dns.Server) {
			d.Logger.WithFields(logrus.Fields{
				"net":  server.Net,
				"addr": server.Addr,
			}).Info("starting dns server")
			errCh <- server.ListenAndServe()
		}(server)
	}

	d.startCachePruner()

	for range d.Servers {
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

	for _, server := range d.Servers {
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
