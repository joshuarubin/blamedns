package dnsserver

import (
	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func (d *DNSServer) Handler(net string) dns.Handler {
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

		d.ValidateDNSSEC(req, resp)
		d.stripRRSIG(req, resp)
		resp.RecursionAvailable = !d.DisableRecursion

		if err = w.WriteMsg(resp); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Error("handler error")
			dns.HandleFailed(w, req)
		}
	})
}

func (d *DNSServer) ValidateDNSSEC(req, resp *dns.Msg) {
	// TODO(jrubin) what to return?

	if d.DisableDNSSEC {
		return
	}

	if req.CheckingDisabled {
		return
	}

	// TODO(jrubin) validate dnssec here
	// TODO(jrubin) if validated, set resp.AuthenticatedData?
}

func (d *DNSServer) stripRRSIG(req, resp *dns.Msg) {
	// TODO(jrubin) handle multiple questions?
	if len(req.Question) > 0 && req.Question[0].Qtype == dns.TypeRRSIG {
		// it was an RRSIG request, don't strip
		return
	}

	if opt := req.IsEdns0(); opt != nil && opt.Do() {
		// DNSSEC OK (+dnssec) was set, keep rrsig values
		return
	}

	var ret []dns.RR
	for _, ans := range resp.Answer {
		if _, ok := ans.(*dns.RRSIG); ok {
			continue
		}
		ret = append(ret, ans)
	}
	resp.Answer = ret
}
