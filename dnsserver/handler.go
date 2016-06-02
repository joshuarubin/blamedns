package dnsserver

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/net/context"

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

func (d *DNSServer) respond(net string, w dns.ResponseWriter, req *dns.Msg, dur time.Duration, r *hresp) *hresp {
	if r == nil {
		r = &hresp{}
	}

	if r.resp == nil {
		r.resp = &dns.Msg{}
		r.resp.SetRcode(req, dns.RcodeServerFailure)
	}

	r.resp.RecursionAvailable = true

	ctxLog := d.Logger.WithFields(logrus.Fields{
		"name":     req.Question[0].Name,
		"type":     dns.TypeToString[req.Question[0].Qtype],
		"net":      net,
		"rcode":    dns.RcodeToString[r.resp.Rcode],
		"blocked":  r.blocked,
		"cache":    r.cache.String(),
		"duration": dur,
	})

	go func() {
		if r.resp.Rcode == dns.RcodeServerFailure {
			ctxLog.Error("responded with error")
		} else if r.blocked {
			ctxLog.Warn("responded")
		} else if r.cache == cacheMiss {
			ctxLog.Info("responded")
		} else {
			ctxLog.Debug("responded")
		}
	}()

	if err := w.WriteMsg(r.resp); err != nil {
		ctxLog.WithError(err).Error("error writing response")
	}

	return r
}

func refused(req *dns.Msg) *hresp {
	resp := &dns.Msg{}
	resp.SetRcode(req, dns.RcodeRefused)
	return &hresp{
		resp:  resp,
		cache: cacheHit,
	}
}

type cacheStatus int

const (
	cacheMiss cacheStatus = iota
	cacheHit
)

func (c cacheStatus) String() string {
	switch c {
	case cacheMiss:
		return "miss"
	case cacheHit:
		return "hit"
	}
	return strconv.Itoa(int(c))
}

type hresp struct {
	resp    *dns.Msg
	blocked bool
	cache   cacheStatus
}

func (d *DNSServer) bgHandler(ctx context.Context, net string, addr []string, req *dns.Msg, respCh chan<- *hresp) {
	// refuse "any" and "rrsig" requests
	switch req.Question[0].Qtype {
	case dns.TypeANY, dns.TypeRRSIG:
		respCh <- refused(req)
		return
	}

	if d.Block.Should(req) {
		respCh <- &hresp{
			resp:    d.Block.NewReply(req),
			blocked: true,
			cache:   cacheHit,
		}
		return
	}

	if d.Cache != nil {
		if resp := d.Cache.Get(ctx, req); resp != nil {
			respCh <- &hresp{
				resp:  resp,
				cache: cacheHit,
			}
			return
		}
	}

	if !req.RecursionDesired {
		respCh <- refused(req)
		return
	}

	respCh <- &hresp{
		resp:  d.fastLookup(ctx, net, addr, req),
		cache: cacheMiss,
	}
}

func (d *DNSServer) Handler(net string, addr []string) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		begin := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), d.DialTimeout+2*d.ClientTimeout)
		respCh := make(chan *hresp, 1)

		go d.bgHandler(ctx, net, addr, req, respCh)

		var r *hresp

		select {
		case <-ctx.Done():
		case r = <-respCh:
			if d.Cache != nil {
				go d.Cache.Set(r.resp)
			}
		}

		cancel()
		dur := time.Since(begin)
		r = d.respond(net, w, req, dur, r)

		handlerDuration.
			WithLabelValues(
				dns.RcodeToString[r.resp.Rcode],
				fmt.Sprintf("%v", r.blocked),
				r.cache.String(),
			).
			Observe(float64(dur))
	})
}
