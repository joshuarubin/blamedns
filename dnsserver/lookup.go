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
func (d DNSServer) Lookup(net string, req *dns.Msg) (*dns.Msg, error) {
	q := req.Question[0]

	logFields := logrus.Fields{
		"name":  unfqdn(q.Name),
		"type":  dns.TypeToString[q.Qtype],
		"class": dns.ClassToString[q.Qclass],
		"net":   net,
	}

	if resp := d.LookupCached(req); resp != nil {
		return resp, nil
	}

	c := &dns.Client{
		Net:          net,
		DialTimeout:  d.DialTimeout,
		ReadTimeout:  d.Timeout,
		WriteTimeout: d.Timeout,
	}

	resCh := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()

		logFields["nameserver"] = nameserver

		res, _, err := c.Exchange(req, nameserver)
		if err != nil {
			d.Logger.WithError(err).WithFields(logFields).Warn("socket error")
			return
		}
		if res != nil && res.Rcode != dns.RcodeSuccess {
			d.Logger.WithFields(logFields).Debugf("failed to get a valid answer")
			if res.Rcode == dns.RcodeServerFailure {
				return
			}
		} else {
			d.Logger.WithFields(logFields).Debug("resolv")
		}

		select {
		case resCh <- res:
		default:
		}
	}

	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range d.Forward {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case res := <-resCh:
			d.StoreCached(res)
			return res, nil
		case <-ticker.C:
			continue
		}
	}

	// wait for all the namservers to finish
	wg.Wait()
	select {
	case res := <-resCh:
		return res, nil
	default:
		return nil, ResolvError{q.Name, net, d.Forward}
	}
}
