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

		if blocked, err := d.Block.IfNeeded(w, req); blocked {
			if err != nil {
				d.Logger.WithError(err).WithFields(logFields).Error("error blocking request")
				return
			}

			d.Logger.WithFields(logFields).Debug("blocked")
			return
		}

		resp, err := d.Lookup(net, req)
		if err != nil {
			d.Logger.WithError(err).Error("lookup error")
			dns.HandleFailed(w, req)
			return
		}

		if err = w.WriteMsg(resp); err != nil {
			d.Logger.WithError(err).Error("handler error")
			dns.HandleFailed(w, req)
		}
	})
}
