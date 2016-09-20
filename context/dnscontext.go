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
			ctx.Block.Start()
			return nil
		},
		Zones:       map[string][]string{},
		OverrideTTL: cfg.DNS.OverrideTTL.Value(),
		Override:    override.New(override.Parse(cfg.DNS.Override)),
		HTTP: dnsserver.DNSHTTP{
			KeepAlive:             cfg.DNS.HTTP.KeepAlive.Value(),
			MaxIdleConns:          cfg.DNS.HTTP.MaxIdleConns,
			MaxIdleConnsPerHost:   cfg.DNS.HTTP.MaxIdleConnsPerHost,
			IdleConnTimeout:       cfg.DNS.HTTP.IdleConnTimeout.Value(),
			TLSHandshakeTimeout:   cfg.DNS.HTTP.TLSHandshakeTimeout.Value(),
			ExpectContinueTimeout: cfg.DNS.HTTP.ExpectContinueTimeout.Value(),
			NoDNSSEC:              cfg.DNS.HTTP.NoDNSSEC,
			EDNSClientSubnet:      cfg.DNS.HTTP.EDNSClientSubnet,
			NoRandomPadding:       cfg.DNS.HTTP.NoRandomPadding,
		},
	}

	if ctx.Cache.Cache != nil {
		ctx.Server.Cache = ctx.Cache.Cache
	}

	if len(cfg.DNS.Forward) > 0 {
		ctx.Server.Zones["."] = cfg.DNS.Forward
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
