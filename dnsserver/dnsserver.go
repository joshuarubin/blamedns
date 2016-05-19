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
	NotifyStartedFunc  func() error
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
		if d.cachePruneTicker == nil {
			// happens when NotifyStartedFunc returns an error and shuts down
			// the server right away
			return
		}

		for range d.cachePruneTicker.C {
			n := d.Cache.Prune()
			d.Logger.Debugf("pruned %d items from cache", n)
		}
	}()
}

func (d *DNSServer) parseDNSServer(val string, startCh chan<- struct{}) (*dns.Server, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}

	return &dns.Server{
		Addr:              u.Host,
		Net:               u.Scheme,
		Handler:           d.Handler(u.Scheme),
		NotifyStartedFunc: func() { startCh <- struct{}{} },
		// TODO(jrubin)
		// ReadTimeout:  2*time.Second,
		// WriteTimeout: 2*time.Second,
	}, nil
}

func (d *DNSServer) createServers(startCh chan<- struct{}) error {
	d.servers = make([]*dns.Server, len(d.Listen))

	for i, listen := range d.Listen {
		server, err := d.parseDNSServer(listen, startCh)
		if err != nil {
			return err
		}
		d.servers[i] = server
	}
	return nil
}

func (d *DNSServer) startedListener() chan<- struct{} {
	startCh := make(chan struct{})
	go func() {
		// wait for each of the servers to report that they started
		for i := 0; i < len(d.Listen); i++ {
			<-startCh
		}
		close(startCh)

		d.startCachePruner()

		if d.NotifyStartedFunc != nil {
			d.Logger.Debug("running NotifyStartedFunc")

			if err := d.NotifyStartedFunc(); err != nil {
				d.Logger.WithError(err).Error("NotifyStartedFunc returned an error, shutting down")
				d.Shutdown()
			}
		}
	}()
	return startCh
}

func (d *DNSServer) ListenAndServe() error {
	var err error
	if err = d.createServers(d.startedListener()); err != nil {
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

	for range d.servers {
		if err := <-errCh; err != nil {
			if nerr, ok := err.(*net.OpError); ok {
				// wish there was a better way to detect this...
				// https://github.com/golang/go/issues/4373
				if nerr.Op == "accept" && nerr.Net == "tcp" && nerr.Err.Error() == "use of closed network connection" {
					// shutdown sometimes causes this... (particularly with
					// NotifyStartedFunc errors)
					d.Logger.Warn(err)
					return nil
				}
			}
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

func (d *DNSServer) Shutdown() {
	d.stopCachePruner()

	for _, server := range d.servers {
		logFields := logrus.Fields{
			"net":  server.Net,
			"addr": server.Addr,
		}
		if err := server.Shutdown(); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Debug("error shutting down dns server")
		} else {
			d.Logger.WithFields(logFields).Debug("successfully shut down dns server")
		}
	}
}
