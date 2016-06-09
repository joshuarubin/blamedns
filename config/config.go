package config

import (
	"fmt"
	"os"
	"runtime"

	"jrubin.io/blamedns/apiserver"
	"jrubin.io/blamedns/pixelserv"
	"jrubin.io/slog"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

type server interface {
	ServerName() string
	ServerAddr() string
	Init() error
	Start() error
	Close() error
}

type Config struct {
	CacheDir        string       `toml:"cache_dir"`
	ListenPixelserv string       `toml:"listen_pixelserv"`
	ListenAPIServer string       `toml:"listen_apiserver"`
	Log             *LogConfig   `toml:"log"`
	DL              *DLConfig    `toml:"dl"`
	DNS             *DNSConfig   `toml:"dns"`
	Logger          *slog.Logger `toml:"-"`
	AppName         string       `toml:"-"`
	AppVersion      string       `toml:"-"`
	servers         []server
}

func defaultCacheDir(name string) string {
	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("%s/Library/Caches/%s", os.Getenv("HOME"), name)
	}

	return fmt.Sprintf("%s/.cache/%s", os.Getenv("HOME"), name)
}

func New(appName, appVersion string) *Config {
	return &Config{
		CacheDir:   defaultCacheDir("blamedns"),
		Log:        NewLogConfig(),
		DL:         NewDLConfig(),
		DNS:        NewDNSConfig(),
		Logger:     slog.New(),
		AppName:    appName,
		AppVersion: appVersion,
	}
}

func (c *Config) Flags() []cli.Flag {
	ret := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        "cache-dir",
			EnvVar:      "CACHE_DIR",
			Value:       c.CacheDir,
			Destination: &c.CacheDir,
			Usage:       "directory in which to store cached data",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        "listen-pixelserv",
			EnvVar:      "LISTEN_PIXELSERV",
			Usage:       "address to run the pixel server on",
			Destination: &c.ListenPixelserv,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        "listen-apiserver",
			EnvVar:      "LISTEN_APISERVER",
			Usage:       "address to run the api server on",
			Destination: &c.ListenAPIServer,
		}),
	}

	ret = append(ret, c.Log.Flags()...)
	ret = append(ret, c.DL.Flags()...)
	ret = append(ret, c.DNS.Flags()...)

	return ret
}

func (c *Config) Get(name string) (interface{}, bool) {
	switch name {
	case "cache-dir":
		return c.CacheDir, true
	case "listen-pixelserv":
		return c.ListenPixelserv, true
	case "listen-apiserver":
		return c.ListenAPIServer, true
	}

	if ret, ok := c.Log.Get(name); ok {
		return ret, ok
	}

	if ret, ok := c.DL.Get(name); ok {
		return ret, ok
	}

	if ret, ok := c.DNS.Get(name); ok {
		return ret, ok
	}

	panic(fmt.Errorf("could not find config value for key %s", name))
}

func (c *Config) Init() error {
	if err := c.Log.Init(c); err != nil {
		return err
	}

	c.servers = []server{
		pixelserv.New(c.ListenPixelserv, c.Logger),
		apiserver.New(c.ListenAPIServer, c.Logger, c.Log.level),
	}

	for _, server := range c.servers {
		if err := server.Init(); err != nil {
			return err
		}
	}

	if err := c.DL.Init(c); err != nil {
		return err
	}

	return c.DNS.Init(c, c.DL.Start)
}

func (c *Config) Start() error {
	for _, server := range c.servers {
		_ = server.Start()
	}

	return c.DNS.Start()
}

func (c *Config) SIGUSR1() {
	c.DNS.SIGUSR1()
}

func (c *Config) Shutdown() {
	c.DL.Shutdown()

	for _, server := range c.servers {
		_ = server.Close()
	}

	c.DNS.Shutdown()
	c.Log.Shutdown()
}
