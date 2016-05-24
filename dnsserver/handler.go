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

func (d *DNSServer) Handler(net string) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		if len(req.Question) == 0 {
			d.RespondCode(net, w, req, dns.RcodeFormatError)
			return
		}

		if len(req.Question) > 1 {
			d.RespondCode(net, w, req, dns.RcodeNotImplemented)
			return
		}

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
		} else if resp, err = d.Lookup(net, req); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Warn("lookup error")
		}

		if resp == nil {
			dns.HandleFailed(w, req)
			return
		}

		d.ValidateDNSSEC(req, resp)
		d.stripRRSIG(req, resp)
		resp.RecursionAvailable = true

		if err = w.WriteMsg(resp); err != nil {
			d.Logger.WithError(err).WithFields(logFields).Error("handler error")
			dns.HandleFailed(w, req)
		}
	})
}

func (d *DNSServer) ValidateDNSSEC(req, resp *dns.Msg) {
	if d.DisableDNSSEC {
		return
	}

	if req.CheckingDisabled {
		return
	}

	// TODO(jrubin) dnssec validation
	// TODO(jrubin)if validated, set resp.AuthenticatedData
}

func (d *DNSServer) stripRRSIG(req, resp *dns.Msg) {
	if req.Question[0].Qtype == dns.TypeRRSIG {
		// it was an RRSIG request, don't strip
		return
	}

	if opt := req.IsEdns0(); opt != nil && opt.Do() {
		// DNSSEC OK (+dnssec) was set, keep rrsig values
		return
	}

	for i, set := range [][]dns.RR{resp.Answer, resp.Ns, resp.Extra} {
		var deleted int
		for j := range set {
			k := j - deleted
			rr := set[k]

			if rr.Header().Rrtype == dns.TypeRRSIG {
				copy(set[k:], set[k+1:])
				set[len(set)-1] = nil
				set = set[:len(set)-1]
				deleted++
			}
		}

		switch i {
		case 0: // Answer
			resp.Answer = set
		case 1: // Ns
			resp.Ns = set
		case 2: // Extra
			resp.Extra = set
		}
	}
}
