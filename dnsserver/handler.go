package dnsserver

import (
	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func (d DNSServer) Handler(net string) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		if len(req.Question) < 1 { // allow more than one question
			dns.HandleFailed(w, req)
			return
		}

		logFields := logrus.Fields{
			"qname": req.Question[0].Name,
			"net":   net,
		}

		var resp *dns.Msg
		var err error

		if d.Block.Should(req) {
			resp = d.Block.NewReply(req)
			d.Logger.WithFields(logFields).Debug("blocked")
		} else if resp, err = d.Lookup(net, req); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Error("lookup error")
		}

		if resp == nil {
			dns.HandleFailed(w, req)
			return
		}

		if err = w.WriteMsg(resp); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Error("handler error")
			dns.HandleFailed(w, req)
		}
	})
}
