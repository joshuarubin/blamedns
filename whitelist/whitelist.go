package whitelist

import "jrubin.io/blamedns/dnsserver"

var _ dnsserver.Passer = &WhiteList{}

type WhiteList map[string]struct{}

func (w WhiteList) Pass(host string) bool {
	_, found := w[host]
	return found
}

func New(domains []string) WhiteList {
	ret := WhiteList{}
	for _, domain := range domains {
		ret[domain] = struct{}{}
	}
	return ret
}
