package dnscache

import (
	"context"

	"github.com/miekg/dns"
)

type Cache interface {
	Get(context.Context, *dns.Msg) *dns.Msg
	Set(*dns.Msg) int
}
