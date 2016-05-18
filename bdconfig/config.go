package bdconfig

import (
	"net"

	"jrubin.io/blamedns/pixelserv"

	"github.com/Sirupsen/logrus"
)

const (
	defaultCacheDir        = "cache"
	defaultListenPixelserv = "[::]:http"
)

type Config struct {
	CacheDir        string         `toml:"cache_dir" cli:",directory in which to store cached data"`
	Log             LogConfig      `toml:"log"`
	DNS             DNSConfig      `toml:"dns"`
	ListenPixelserv string         `toml:"listen_pixelserv" cli:",address to run the pixel server on"`
	Logger          *logrus.Logger `toml:"-" cli:"-"`
	AppName         string         `toml:"-" cli:"-"`
	AppVersion      string         `toml:"-" cli:"-"`
}

func New(appName, appVersion string) *Config {
	return &Config{
		Logger:     logrus.New(),
		AppName:    appName,
		AppVersion: appVersion,
	}
}

func Default() Config {
	return Config{
		CacheDir:        defaultCacheDir,
		Log:             defaultLogConfig(),
		DNS:             defaultDNSConfig(),
		ListenPixelserv: defaultListenPixelserv,
	}
}

func (m *Config) Init() error {
	m.Log.Init(m)
	return m.DNS.Init(m)
}

func (m Config) Start() error {
	l, err := net.Listen("tcp", m.ListenPixelserv)
	if err != nil {
		return err
	}
	defer l.Close()
	m.Logger.WithFields(logrus.Fields{
		"Addr": m.ListenPixelserv,
		"Net":  "tcp",
	}).Info("starting pixelserv")
	go pixelserv.Serve(l)

	return m.DNS.Start()
}

func (m Config) Shutdown() error {
	return m.DNS.Shutdown()
}
