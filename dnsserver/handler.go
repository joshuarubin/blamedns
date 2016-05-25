package dnsserver

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	prom "github.com/prometheus/client_golang/prometheus"
)

var handlerDuration = prom.NewSummaryVec(
	prom.SummaryOpts{
		Namespace: "blamedns",
		Subsystem: "dns",
		Name:      "handler_duration_nanoseconds",
		Help:      "Total time spent serving dns requests.",
	},
	[]string{"rcode", "blocked", "cache"},
)

func init() {
	prom.MustRegister(handlerDuration)
}

func observeHandler(begin time.Time, resp **dns.Msg, blocked *bool, cache *string) {
	rcode := dns.RcodeServerFailure
	if resp != nil && *resp != nil {
		rcode = (*resp).Rcode
	}

	handlerDuration.
		WithLabelValues(dns.RcodeToString[rcode], fmt.Sprintf("%v", *blocked), *cache).
		Observe(float64(time.Since(begin)))
}

func (d *DNSServer) respond(net string, w dns.ResponseWriter, req *dns.Msg, resp **dns.Msg, blocked *bool, cache *string) {
	if resp == nil || *resp == nil {
		tmp := &dns.Msg{}
		tmp.SetRcode(req, dns.RcodeServerFailure)
		resp = &tmp
	}

	(*resp).RecursionAvailable = true

	lf := logFields(net, req, *resp, blocked, cache)

	if (*resp).Rcode == dns.RcodeServerFailure {
		d.Logger.WithFields(lf).Error("responded with error")
	} else {
		d.Logger.WithFields(lf).Debug("responded")
	}

	if err := w.WriteMsg(*resp); err != nil {
		d.Logger.WithError(err).WithFields(lf).Error("error writing response")
	}
}

func refused(req *dns.Msg) *dns.Msg {
	resp := &dns.Msg{}
	resp.SetRcode(req, dns.RcodeRefused)
	return resp
}

func logFields(net string, req, resp *dns.Msg, blocked *bool, cache *string) logrus.Fields {
	ret := logrus.Fields{
		"name": req.Question[0].Name,
		"type": dns.TypeToString[req.Question[0].Qtype],
		"net":  net,
	}

	if resp != nil {
		ret["rcode"] = dns.RcodeToString[(*resp).Rcode]
	}

	if blocked != nil {
		ret["blocked"] = *blocked
	}

	if cache != nil {
		ret["cache"] = *cache
	}

	return ret
}

func (d *DNSServer) Handler(net string, addr []string) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		var resp *dns.Msg
		blocked := false
		cache := "miss"

		defer d.respond(net, w, req, &resp, &blocked, &cache)
		defer observeHandler(time.Now(), &resp, &blocked, &cache)

		// refuse "any" and "rrsig" requests
		switch req.Question[0].Qtype {
		case dns.TypeANY, dns.TypeRRSIG:
			resp = refused(req)
			return
		}

		if d.Block.Should(req) {
			blocked = true
			cache = "hit"
			resp = d.Block.NewReply(req)
			return
		}

		if d.Cache != nil {
			if resp = d.Cache.Get(req); resp != nil {
				cache = "hit"
				return
			}
		}

		if !req.RecursionDesired {
			resp = refused(req)
			return
		}

		resp = d.fastLookup(net, addr, req)

		if d.Cache != nil {
			go d.Cache.Set(resp)
		}
	})
}
