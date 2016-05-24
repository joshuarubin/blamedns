package bdconfig

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"jrubin.io/blamedns/pixelserv"

	"github.com/Sirupsen/logrus"
)

const defaultListenPixelserv = "[::]:http"

func defaultCacheDir(name string) string {
	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("%s/Library/Caches/%s", os.Getenv("HOME"), name)
	}

	return fmt.Sprintf("%s/.cache/%s", os.Getenv("HOME"), name)
}

type Config struct {
	CacheDir        string            `toml:"cache_dir" cli:",directory in which to store cached data"`
	Log             LogConfig         `toml:"log"`
	DL              DLConfig          `toml:"dl"`
	DNS             DNSConfig         `toml:"dns"`
	ListenPixelserv string            `toml:"listen_pixelserv" cli:",address to run the pixel server on"`
	Logger          *logrus.Logger    `toml:"-" cli:"-"`
	AppName         string            `toml:"-" cli:"-"`
	AppVersion      string            `toml:"-" cli:"-"`
	pixelServ       *pixelserv.Server `toml:"-" cli:"-"`
}

func New(appName, appVersion string) *Config {
	return &Config{
		Logger:     logrus.New(),
		AppName:    appName,
		AppVersion: appVersion,
		DNS: DNSConfig{
			ForwardZone: defaultDNSForwardZones,
		},
	}
}

func Default(appName, appVersion string) Config {
	return Config{
		CacheDir:        defaultCacheDir(appName),
		Log:             defaultLogConfig,
		DL:              defaultDLConfig,
		DNS:             defaultDNSConfig,
		ListenPixelserv: defaultListenPixelserv,
		AppName:         appName,
		AppVersion:      appVersion,
	}
}

func (m *Config) Init() error {
	m.Log.Init(m)
	m.pixelServ = &pixelserv.Server{Logger: m.Logger}
	if err := m.DL.Init(m); err != nil {
		return err
	}
	return m.DNS.Init(m, m.DL.Start)
}

func (m Config) Start() error {
	l, err := net.Listen("tcp", m.ListenPixelserv)
	if err != nil {
		return err
	}
	defer func() { _ = l.Close() }()
	go func() { _ = m.pixelServ.Serve(l) }()

	return m.DNS.Start()
}

func (m Config) Shutdown() {
	m.DL.Shutdown()
	m.DNS.Shutdown()
}
