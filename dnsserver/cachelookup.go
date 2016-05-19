package dnsserver

import "github.com/miekg/dns"

func (d DNSServer) LookupCached(req *dns.Msg) *dns.Msg {
	if d.Cache == nil {
		return nil
	}

	// TODO(jrubin)

	return nil
}

func (d DNSServer) StoreCached(res *dns.Msg) {
	if d.Cache == nil {
		return
	}

	d.Cache.Set(res)
}
