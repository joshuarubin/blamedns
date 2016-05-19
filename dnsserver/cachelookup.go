package dnsserver

import (
	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func logFields(q dns.Question, cache string) logrus.Fields {
	return logrus.Fields{
		"class": dns.ClassToString[q.Qclass],
		"type":  dns.TypeToString[q.Qtype],
		"name":  unfqdn(q.Name),
		"cache": cache,
	}
}

func (d DNSServer) LookupCached(req *dns.Msg) *dns.Msg {
	if d.Cache == nil {
		return nil
	}

	// TODO(jrubin) handle multiple questions?
	q := req.Question[0]

	ans := d.Cache.Get(q)
	if ans == nil {
		d.Logger.WithFields(logFields(q, "miss")).Debug("cache lookup")
		return nil
	}

	d.Logger.WithFields(logFields(q, "hit")).Debug("cache lookup")

	resp := &dns.Msg{}
	resp.SetReply(req)
	resp.Answer = ans

	if !d.DisableDNSSEC {
		// add rrsig records (regardless of DNSSEC OK flag)
		q = req.Copy().Question[0]
		q.Qtype = dns.TypeRRSIG

		var lf logrus.Fields
		if ans = d.Cache.Get(q); ans != nil {
			lf = logFields(q, "hit")
			resp.Answer = append(resp.Answer, ans...)
		} else {
			lf = logFields(q, "miss")
		}

		d.Logger.WithFields(lf).Debug("cache lookup (rrsig)")
	}

	d.processAdditionalSection(resp)

	// TODO(jrubin) ensure no duplicates of resp.Answer are in resp.Extra

	return resp
}

func additionalQ(qs []dns.Question, name string) []dns.Question {
	for _, t := range []uint16{dns.TypeA, dns.TypeAAAA} {
		qs = append(qs, dns.Question{
			Name:   name,
			Qtype:  t,
			Qclass: dns.ClassINET,
		})
	}
	return qs
}

func (d DNSServer) processAdditionalSection(resp *dns.Msg) {
	// CNAMEs do not cause "additional section processing"
	// https://tools.ietf.org/html/rfc2181#section-10.3
	//
	// Not Implemented:
	// MB (experimental)
	// MD (obsolete)
	// MF (obsolete)
	// Sig?

	var qs []dns.Question

	for _, rr := range resp.Answer {
		switch t := rr.(type) {
		case *dns.NS:
			qs = additionalQ(qs, t.Ns)
		case *dns.MX:
			qs = additionalQ(qs, t.Mx)
		case *dns.SRV:
			qs = additionalQ(qs, t.Target)
		}
	}

	for _, q := range qs {
		var lf logrus.Fields
		if extra := d.Cache.Get(q); extra != nil {
			lf = logFields(q, "hit")
			resp.Extra = append(resp.Extra, extra...)
		} else {
			lf = logFields(q, "miss")
		}

		d.Logger.WithFields(lf).Debug("cache lookup (additional)")
	}
}

func (d DNSServer) StoreCached(resp *dns.Msg) {
	if d.Cache == nil {
		return
	}

	// TODO(jrubin) negative response caching? (based on SOA TTL)

	d.Cache.Set(resp)
}
