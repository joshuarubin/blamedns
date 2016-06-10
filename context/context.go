package context

import (
	"jrubin.io/blamedns/apiserver"
	"jrubin.io/blamedns/config"
	"jrubin.io/blamedns/pixelserv"
)

type Context struct {
	AppName    string
	AppVersion string
	Log        *LogContext
	DL         *DLContext
	DNS        *DNSContext
	servers    []server
}

type server interface {
	ServerName() string
	ServerAddr() string
	Init() error
	Start() error
	Close() error
}

func New(appName, appVersion string, cfg *config.Config) (*Context, error) {
	logCtx, err := NewLogContext(cfg.Log)
	if err != nil {
		return nil, err
	}

	ctx := &Context{
		AppName:    appName,
		AppVersion: appVersion,
		Log:        logCtx,
	}

	ctx.servers = []server{
		pixelserv.New(cfg.ListenPixelserv, ctx.Log.Logger),
		apiserver.New(cfg.ListenAPIServer, ctx.Log.Logger, ctx.Log.Level),
	}

	for _, server := range ctx.servers {
		if err = server.Init(); err != nil {
			return nil, err
		}
	}

	if ctx.DL, err = NewDLContext(ctx, cfg); err != nil {
		return nil, err
	}

	if ctx.DNS, err = NewDNSContext(ctx.Log.Logger, cfg, ctx.DL.Start); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (ctx *Context) Start() error {
	for _, server := range ctx.servers {
		_ = server.Start()
	}

	return ctx.DNS.Start()
}

func (ctx *Context) SIGUSR1() {
	ctx.DNS.SIGUSR1()
}

func (ctx *Context) Shutdown() {
	ctx.DL.Shutdown()

	for _, server := range ctx.servers {
		_ = server.Close()
	}

	ctx.DNS.Shutdown()
	ctx.Log.Shutdown()
}
