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
func (d *DNSServer) Lookup(net string, req *dns.Msg) (*dns.Msg, error) {
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

	if !req.RecursionDesired || d.DisableRecursion {
		resp := &dns.Msg{}
		resp.SetRcode(req, dns.RcodeRefused)
		return resp, nil
	}

	// TODO(jrubin) handle recursive requests without forward servers
	// will need root.hints
	// if len(d.Forward) == 0 {
	// for n, i := 1, len(q.Name); i > 0; n++ {
	// 	i, _ = dns.PrevLabel(q.Name, n)
	// 	label := q.Name[i:]
	// 	fmt.Println(i, label)
	// }
	// }

	c := &dns.Client{
		Net:          net,
		DialTimeout:  d.DialTimeout,
		ReadTimeout:  d.ClientTimeout,
		WriteTimeout: d.ClientTimeout,
	}

	freq := req.Copy()

	if !d.DisableDNSSEC {
		// ensure all forward requests request dnssec
		if opt := freq.IsEdns0(); opt == nil || !opt.Do() {
			// only set it if it wasn't already set
			freq.SetEdns0(4096, true)
		}
	}

	resCh := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()

		logFields["nameserver"] = nameserver

		resp, _, err := c.Exchange(freq, nameserver)
		if err != nil {
			d.Logger.WithError(err).WithFields(logFields).Warn("socket error")
			return
		}
		if resp != nil && resp.Rcode != dns.RcodeSuccess {
			d.Logger.WithFields(logFields).Debugf("failed to get a valid answer")
			if resp.Rcode == dns.RcodeServerFailure {
				return
			}
		} else {
			d.Logger.WithFields(logFields).Debug("resolv")
		}

		select {
		case resCh <- resp:
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
		case resp := <-resCh:
			// TODO(jrubin) validate dnssec here and only cache if valid?
			d.StoreCached(resp)
			return resp, nil
		case <-ticker.C:
			continue
		}
	}

	// wait for all the namservers to finish
	wg.Wait()
	select {
	case resp := <-resCh:
		return resp, nil
	default:
		return nil, ResolvError{q.Name, net, d.Forward}
	}
}
