package bdconfig

import (
	"fmt"
	"os"
	"runtime"

	"jrubin.io/blamedns/apiserver"
	"jrubin.io/blamedns/pixelserv"

	"github.com/Sirupsen/logrus"
)

func defaultCacheDir(name string) string {
	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("%s/Library/Caches/%s", os.Getenv("HOME"), name)
	}

	return fmt.Sprintf("%s/.cache/%s", os.Getenv("HOME"), name)
}

type server interface {
	ServerName() string
	ServerAddr() string
	Init() error
	Start() error
	Close() error
}

type Config struct {
	CacheDir        string         `toml:"cache_dir" cli:",directory in which to store cached data"`
	Log             LogConfig      `toml:"log"`
	DL              DLConfig       `toml:"dl"`
	DNS             DNSConfig      `toml:"dns"`
	ListenPixelserv string         `toml:"listen_pixelserv" cli:",address to run the pixel server on"`
	ListenAPIServer string         `toml:"listen_api_server" cli:",address to run the api server on"`
	Logger          *logrus.Logger `toml:"-" cli:"-"`
	AppName         string         `toml:"-" cli:"-"`
	AppVersion      string         `toml:"-" cli:"-"`
	servers         []server
}

func New(appName, appVersion string) *Config {
	return &Config{
		Logger:     logrus.New(),
		AppName:    appName,
		AppVersion: appVersion,
	}
}

func Default(appName, appVersion string) Config {
	return Config{
		CacheDir:   defaultCacheDir(appName),
		Log:        defaultLogConfig,
		DL:         defaultDLConfig,
		DNS:        defaultDNSConfig,
		AppName:    appName,
		AppVersion: appVersion,
	}
}

func (m *Config) Init() error {
	m.Log.Init(m)

	m.servers = []server{
		pixelserv.New(m.ListenPixelserv, m.Logger),
		apiserver.New(m.ListenAPIServer, m.Logger),
	}

	for _, server := range m.servers {
		if err := server.Init(); err != nil {
			return err
		}
	}

	if err := m.DL.Init(m); err != nil {
		return err
	}
	return m.DNS.Init(m, m.DL.Start)
}

func (m *Config) Start() error {
	for _, server := range m.servers {
		_ = server.Start()
	}

	return m.DNS.Start()
}

func (m *Config) Shutdown() {
	for _, server := range m.servers {
		_ = server.Close()
	}
	m.DL.Shutdown()
	m.DNS.Shutdown()
	m.Log.Shutdown()
}

func (m *Config) SIGUSR1() {
	m.DNS.SIGUSR1()
}
