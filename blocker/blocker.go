package blocker

import (
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
)

type Blocker interface {
	parser.HostAdder
	dnsserver.Blocker
	Len() int
}
