package override

import (
	"net"

	"github.com/armon/go-radix"
	"jrubin.io/blamedns/parser"
)

type Override struct {
	data *radix.Tree
}

func copyIPs(in []net.IP) []net.IP {
	// wow this is ugly...
	ret := make([]net.IP, len(in))
	for i, ip := range in {
		r := make(net.IP, len(ip))
		copy(r, ip)
		ret[i] = r
	}

	return ret
}

func Parse(data map[string][]string) map[string][]net.IP {
	ret := make(map[string][]net.IP, len(data))

	for host, ips := range data {
		ret[host] = ParseIPs(ips...)
	}

	return ret
}

func ParseIPs(in ...string) []net.IP {
	ret := make([]net.IP, len(in))

	for i, ip := range in {
		ret[i] = net.ParseIP(ip)
	}

	return ret
}

func New(data map[string][]net.IP) *Override {
	ret := &Override{
		data: radix.New(),
	}

	for host, ips := range data {
		key := parser.ReverseHostName(host)
		ret.data.Insert(key, copyIPs(ips))
	}

	return ret
}

func (o Override) Override(host string) []net.IP {
	key := parser.ReverseHostName(host)
	val, ok := o.data.Get(key)
	if !ok {
		return nil
	}

	v, ok := val.([]net.IP)
	if !ok {
		return nil
	}

	return copyIPs(v)
}
