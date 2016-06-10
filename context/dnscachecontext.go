package context

import (
	"time"

	"jrubin.io/blamedns/config"
	"jrubin.io/blamedns/dnscache"
	"jrubin.io/slog"
)

type DNSCacheContext struct {
	Cache         *dnscache.Memory
	PruneInterval time.Duration
}

func NewDNSCacheContext(logger slog.Interface, cfg *config.DNSCacheConfig) *DNSCacheContext {
	ctx := &DNSCacheContext{
		PruneInterval: cfg.PruneInterval.Duration(),
	}
	if !cfg.Disable {
		ctx.Cache = dnscache.NewMemory(cfg.Size, logger)
	}
	return ctx
}

func (ctx DNSCacheContext) Start() {
	if ctx.Cache != nil {
		ctx.Cache.Start(ctx.PruneInterval)
	}
}

func (ctx *DNSCacheContext) SIGUSR1() {
	if ctx.Cache != nil {
		ctx.Cache.Purge()
	}
}

func (ctx DNSCacheContext) Shutdown() {
	if ctx.Cache != nil {
		ctx.Cache.Stop()
	}
}
