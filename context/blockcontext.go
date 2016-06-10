package context

import (
	"path"

	"jrubin.io/blamedns/blocker"
	"jrubin.io/blamedns/config"
	"jrubin.io/blamedns/dnsserver"
	"jrubin.io/blamedns/parser"
	"jrubin.io/blamedns/watcher"
	"jrubin.io/blamedns/whitelist"
	"jrubin.io/slog"
)

type BlockContext struct {
	Watchers []*watcher.Watcher
	Block    dnsserver.Block
}

func NewBlockContext(logger *slog.Logger, cfg *config.Config) (*BlockContext, error) {
	hostsDir := path.Join(cfg.CacheDir, "hosts")
	domainsDir := path.Join(cfg.CacheDir, "domains")

	blocker := &blocker.RadixBlocker{}

	ctx := &BlockContext{
		Block: dnsserver.Block{
			IPv4:    cfg.DNS.Block.IPv4.Value(),
			IPv6:    cfg.DNS.Block.IPv6.Value(),
			TTL:     cfg.DNS.Block.TTL.Value(),
			Blocker: blocker,
			Passer:  whitelist.New(cfg.DNS.Block.WhiteList...),
			Logger:  logger,
		},
	}

	hostsFileParser := parser.HostsFileParser{
		HostAdder: blocker,
		Logger:    logger,
	}

	hostsWatcher, err := watcher.New(logger, hostsFileParser, hostsDir)
	if err != nil {
		return nil, err
	}

	domainParser := parser.DomainParser{
		HostAdder: blocker,
		Logger:    logger,
	}

	domainsWatcher, err := watcher.New(logger, domainParser, domainsDir)
	if err != nil {
		return nil, err
	}

	ctx.Watchers = []*watcher.Watcher{
		hostsWatcher,
		domainsWatcher,
	}

	return ctx, nil
}

func (ctx *BlockContext) Start() error {
	for _, w := range ctx.Watchers {
		w.Start()
	}

	return nil
}

func (ctx *BlockContext) Shutdown() {
	for _, w := range ctx.Watchers {
		w.Stop()
	}
}
