package bdconfig

import (
	"fmt"
	"net"
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

type Config struct {
	CacheDir        string            `toml:"cache_dir" cli:",directory in which to store cached data"`
	Log             LogConfig         `toml:"log"`
	DL              DLConfig          `toml:"dl"`
	DNS             DNSConfig         `toml:"dns"`
	ListenPixelserv string            `toml:"listen_pixelserv" cli:",address to run the pixel server on"`
	ListenApiServer string            `toml:"listen_api_server" cli:",address to run the api server on"`
	Logger          *logrus.Logger    `toml:"-" cli:"-"`
	AppName         string            `toml:"-" cli:"-"`
	AppVersion      string            `toml:"-" cli:"-"`
	pixelServ       *pixelserv.Server `toml:"-" cli:"-"`
	apiServer       *apiserver.Server `toml:"-" cli:"-"`
	closers         []closer          `toml:"-" cli:"-"`
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
	m.pixelServ = &pixelserv.Server{Logger: m.Logger}
	m.apiServer = &apiserver.Server{Logger: m.Logger}
	if err := m.DL.Init(m); err != nil {
		return err
	}
	return m.DNS.Init(m, m.DL.Start)
}

type webServer struct {
	serve    func(net.Listener) error
	addr     string
	name     string
	listener net.Listener
}

func (s webServer) Close() error {
	return s.listener.Close()
}

func (s webServer) Name() string {
	return s.name
}

func (s webServer) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *webServer) Start() error {
	if len(s.addr) == 0 {
		return nil
	}

	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go func() { _ = s.serve(s.listener) }()
	return nil
}

func (m Config) webServers() []webServer {
	return []webServer{{
		serve: m.pixelServ.Serve,
		addr:  m.ListenPixelserv,
		name:  "pixelserv",
	}, {
		serve: m.apiServer.Serve,
		addr:  m.ListenApiServer,
		name:  "apiserver",
	}}
}

func (m *Config) startWebServer(s webServer) error {
	if err := s.Start(); err != nil {
		return err
	}

	if s.listener != nil {
		m.closers = append(m.closers, s)
	}

	return nil
}

func (m *Config) close() {
	for _, closer := range m.closers {
		lf := logrus.Fields{
			"server": closer.Name(),
			"addr":   closer.Addr(),
		}

		m.Logger.WithFields(lf).Debug("closing")
		if err := closer.Close(); err != nil {
			m.Logger.WithError(err).WithFields(lf).Warn("error closing")
		}
	}
	m.closers = nil
}

type closer interface {
	Close() error
	Name() string
	Addr() net.Addr
}

func (m *Config) Start() error {
	for _, s := range m.webServers() {
		m.startWebServer(s)
	}

	return m.DNS.Start()
}

func (m *Config) Shutdown() {
	m.close()
	m.DL.Shutdown()
	m.DNS.Shutdown()
	m.Log.Shutdown()
}
