package whitelist

import (
	"github.com/armon/go-radix"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
)

var _ dnsserver.Passer = &WhiteList{}

type WhiteList struct {
	data *radix.Tree
}

func (w WhiteList) Pass(host string) bool {
	key := parser.ReverseHostName(host)
	_, ok := w.data.Get(key)
	return ok
}

func New(domains ...string) WhiteList {
	ret := WhiteList{
		data: radix.New(),
	}

	for _, domain := range domains {
		key := parser.ReverseHostName(domain)
		ret.data.Insert(key, nil)
	}

	return ret
}
