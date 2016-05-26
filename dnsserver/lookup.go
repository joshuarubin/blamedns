package dnsserver

import (
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/miekg/dns"
)

// TODO use context
func (d *DNSServer) fastLookup(ctx context.Context, net string, addr []string, req *dns.Msg) (resp *dns.Msg) {
	respCh := make(chan *dns.Msg)
	var wg sync.WaitGroup

	ticker := time.NewTicker(d.LookupInterval)
	defer ticker.Stop()

	// start lookup on each nameserver top-down, every LookupInterval
	for _, nameserver := range addr {
		wg.Add(1)
		go d.lookup(net, nameserver, req, respCh, &wg)

		// but exit early, if we have an answer
		select {
		case resp = <-respCh:
			return
		case <-ticker.C:
			continue
		}
	}

	wg.Wait() // wait for all the namservers to finish

	select {
	case resp = <-respCh:
		return
	default:
		return
	}
}

func (d *DNSServer) lookup(net, nameserver string, req *dns.Msg, respCh chan<- *dns.Msg, wg *sync.WaitGroup) {
	defer wg.Done()

	c := &dns.Client{
		Net:          net,
		DialTimeout:  d.DialTimeout,
		ReadTimeout:  d.ClientTimeout,
		WriteTimeout: d.ClientTimeout,
	}

	resp, _, err := c.Exchange(req, nameserver)
	if err != nil {
		lf := logFields(net, req, nil)
		lf["nameserver"] = nameserver

		d.Logger.WithError(err).WithFields(lf).Warn("socket error")
		return
	}

	if resp != nil && resp.Rcode != dns.RcodeSuccess && resp.Rcode != dns.RcodeNameError {
		lf := logFields(net, req, nil)
		lf["nameserver"] = nameserver

		d.Logger.WithFields(lf).Warn("failed to get a valid answer")
		if resp.Rcode == dns.RcodeServerFailure {
			return
		}
	}

	select {
	case respCh <- resp:
	default:
	}
}
