package context

import (
	"net/url"
	"path"

	"jrubin.io/blamedns/config"
	"jrubin.io/blamedns/dl"
)

type DLContext struct {
	dl []*dl.DL
}

func NewDLContext(rootCtx *Context, cfg *config.Config) (*DLContext, error) {
	ctx := &DLContext{}
	hostsDir := path.Join(cfg.CacheDir, "hosts")
	domainsDir := path.Join(cfg.CacheDir, "domains")

	for _, t := range []struct {
		Values  []string
		BaseDir string
	}{{
		Values:  cfg.DL.Hosts,
		BaseDir: hostsDir,
	}, {
		Values:  cfg.DL.Domains,
		BaseDir: domainsDir,
	}} {
		for _, u := range t.Values {
			p, err := url.Parse(u)
			if err != nil {
				return nil, err
			}

			d := &dl.DL{
				URL:            p,
				BaseDir:        t.BaseDir,
				UpdateInterval: cfg.DL.UpdateInterval.Duration(),
				Logger:         rootCtx.Log.Logger,
				AppName:        rootCtx.AppName,
				AppVersion:     rootCtx.AppVersion,
				DebugHTTP:      cfg.DL.DebugHTTP,
			}

			if err = d.Init(); err != nil {
				return nil, err
			}

			ctx.dl = append(ctx.dl, d)
		}
	}

	return ctx, nil
}

func (ctx *DLContext) Start() {
	for _, d := range ctx.dl {
		d.Start()
	}
}

func (ctx *DLContext) Shutdown() {
	for _, d := range ctx.dl {
		d.Stop()
	}
}
