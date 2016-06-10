package context

import (
	"jrubin.io/blamedns/config"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/override"
	"jrubin.io/slog"
)

type DNSContext struct {
	Server *dnsserver.DNSServer
	Block  *BlockContext
	Cache  *DNSCacheContext
}

func NewDNSContext(logger slog.Interface, cfg *config.Config, onStart func()) (*DNSContext, error) {
	blockContext, err := NewBlockContext(logger, cfg)
	if err != nil {
		return nil, err
	}

	ctx := &DNSContext{
		Block: blockContext,
		Cache: NewDNSCacheContext(logger, cfg.DNS.Cache),
	}

	ctx.Server = &dnsserver.DNSServer{
		Listen:         cfg.DNS.Listen,
		Block:          blockContext.Block,
		ClientTimeout:  cfg.DNS.ClientTimeout.Value(),
		ServerTimeout:  cfg.DNS.ServerTimeout.Value(),
		DialTimeout:    cfg.DNS.DialTimeout.Value(),
		LookupInterval: cfg.DNS.LookupInterval.Value(),
		Logger:         logger.WithField("system", "dns"),
		NotifyStartedFunc: func() error {
			ctx.Cache.Start()
			if onStart != nil {
				onStart()
			}
			return ctx.Block.Start()
		},
		Cache: ctx.Cache.Cache,
		Zones: map[string][]string{
			".": cfg.DNS.Forward,
		},
		OverrideTTL: cfg.DNS.OverrideTTL.Value(),
		Override:    override.New(override.Parse(cfg.DNS.Override)),
	}

	for _, zone := range cfg.DNS.Zone {
		ctx.Server.Zones[zone.Name] = zone.Addr
	}

	return ctx, nil
}

func (ctx DNSContext) Start() error {
	// ctx.Block and ctx.Cache are started by DNSServer.NotifyStartedFunc
	return ctx.Server.ListenAndServe()
}

func (ctx DNSContext) SIGUSR1() {
	ctx.Cache.SIGUSR1()
}

func (ctx DNSContext) Shutdown() {
	ctx.Cache.Shutdown()
	ctx.Block.Shutdown()
	ctx.Server.Shutdown()
}
