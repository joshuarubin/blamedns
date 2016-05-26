package dnscache

import (
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type Cache interface {
	Get(context.Context, *dns.Msg) *dns.Msg
	Set(*dns.Msg) int
}
