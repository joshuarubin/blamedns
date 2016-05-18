package hosts

import "fmt"

type WhiteList map[string]struct{}

func (w WhiteList) Pass(host string) bool {
	_, found := w[host]
	return found
}

func NewWhiteList(domains []string) WhiteList {
	ret := WhiteList{}
	for _, domain := range domains {
		fmt.Println("adding to whitelist", domain)
		ret[domain] = struct{}{}
	}
	return ret
}
