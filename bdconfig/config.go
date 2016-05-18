package bdconfig

import "github.com/Sirupsen/logrus"

type Config struct {
	CacheDir   string         `toml:"cache_dir" cli:",directory in which to store cached data"`
	Log        LogConfig      `toml:"log"`
	DNS        DNSConfig      `toml:"dns"`
	Logger     *logrus.Logger `toml:"-" cli:"-"`
	AppName    string         `toml:"-" cli:"-"`
	AppVersion string         `toml:"-" cli:"-"`
}

func New(appName, appVersion string) *Config {
	return &Config{
		Logger:     logrus.New(),
		AppName:    appName,
		AppVersion: appVersion,
	}
}

const defaultCacheDir = "cache"

func Default() Config {
	return Config{
		CacheDir: defaultCacheDir,
		Log:      defaultLogConfig(),
		DNS:      defaultDNSConfig(),
	}
}

func (m *Config) Init() error {
	m.Log.Init(m)
	return m.DNS.Init(m)
}

func (m Config) Start() error {
	return m.DNS.Start()
}

func (m Config) Shutdown() error {
	return m.DNS.Shutdown()
}
