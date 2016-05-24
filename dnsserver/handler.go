package dnsserver

import (
	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func (d *DNSServer) RespondCode(net string, w dns.ResponseWriter, req *dns.Msg, code int) {
	logFields := logrus.Fields{
		"name":  req.Question[0].Name,
		"type":  dns.TypeToString[req.Question[0].Qtype],
		"net":   net,
		"rcode": dns.RcodeToString[code],
	}

	d.Logger.WithFields(logFields).Debug("responding with rcode")

	resp := &dns.Msg{}
	resp.SetRcode(req, code)
	if err := w.WriteMsg(resp); err != nil {
		d.Logger.WithError(err).WithFields(logFields).Error("handler error")
		dns.HandleFailed(w, req)
	}
}

func (d *DNSServer) Handler(net string, addr []string) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		// refuse "any" and "rrsig" requests
		switch req.Question[0].Qtype {
		case dns.TypeANY, dns.TypeRRSIG:
			d.RespondCode(net, w, req, dns.RcodeRefused)
			return
		}

		logFields := logrus.Fields{
			"name": req.Question[0].Name,
			"type": dns.TypeToString[req.Question[0].Qtype],
			"net":  net,
		}

		var resp *dns.Msg
		var err error

		if d.Block.Should(req) {
			resp = d.Block.NewReply(req)
			d.Logger.WithFields(logFields).Debug("blocked")
		} else if resp, err = d.Lookup(net, addr, req); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Warn("lookup error")
		}

		if resp == nil {
			dns.HandleFailed(w, req)
			return
		}

		resp.RecursionAvailable = true

		if err = w.WriteMsg(resp); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Error("handler error")
			dns.HandleFailed(w, req)
		}
	})
}
