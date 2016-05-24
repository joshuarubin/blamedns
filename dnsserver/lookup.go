package dnsserver

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

// Lookup will ask each nameserver in top-to-bottom fashion, starting a new request
// in every second, and return as early as possbile (have an answer).
// It returns an error if no request has succeeded.
//
// largely taken from github.com/looterz/grimd
func (d *DNSServer) Lookup(net string, addr []string, req *dns.Msg) (resp *dns.Msg, err error) {
	if d.Cache != nil {
		if resp = d.Cache.Get(req); resp != nil {
			return
		}
	}

	if !req.RecursionDesired {
		resp = &dns.Msg{}
		resp.SetRcode(req, dns.RcodeRefused)
		return
	}

	defer func() {
		if err == nil && d.Cache != nil {
			go d.Cache.Set(resp)
		}
	}()

	if len(addr) > 0 {
		resp, err = d.fastLookup(net, addr, req)
		if err == nil {
			return
		}
	}

	// TODO(jrubin) resolve recursive requests using root.hints (if no addr)
	err = ResolvError{req.Question[0].Name, net, addr}
	return
}

func (d *DNSServer) fastLookup(net string, addr []string, req *dns.Msg) (resp *dns.Msg, err error) {
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
		err = ResolvError{req.Question[0].Name, net, addr}
		return
	}
}

func (d *DNSServer) lookup(net, nameserver string, req *dns.Msg, respCh chan<- *dns.Msg, wg *sync.WaitGroup) {
	defer wg.Done()

	q := req.Question[0]

	logFields := logrus.Fields{
		"name":       unfqdn(q.Name),
		"type":       dns.TypeToString[q.Qtype],
		"class":      dns.ClassToString[q.Qclass],
		"net":        net,
		"nameserver": nameserver,
	}

	c := &dns.Client{
		Net:          net,
		DialTimeout:  d.DialTimeout,
		ReadTimeout:  d.ClientTimeout,
		WriteTimeout: d.ClientTimeout,
	}

	resp, _, err := c.Exchange(req, nameserver)
	if err != nil {
		d.Logger.WithError(err).WithFields(logFields).Warn("socket error")
		return
	}

	if resp != nil && resp.Rcode != dns.RcodeSuccess && resp.Rcode != dns.RcodeNameError {
		d.Logger.WithFields(logFields).Debugf("failed to get a valid answer")
		if resp.Rcode == dns.RcodeServerFailure {
			return
		}
	} else {
		d.Logger.WithFields(logFields).Debug("resolv")
	}

	select {
	case respCh <- resp:
	default:
	}
}
