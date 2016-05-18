package bdconfig

import (
	"fmt"
	"io"
)

type Config struct {
	CacheDir string `toml:"cache_dir" cli:",directory in which to store cached data"`
	Log      LogConfig
	DNS      DNSConfig
}

const defaultCacheDir = "cache"

func (m Config) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "cache_dir = \"%s\"\n", m.CacheDir)
	if err != nil {
		return n, err
	}

	o, err := m.Log.Write(w)
	n += o
	if err != nil {
		return n, err
	}

	o, err = m.DNS.Write(w)
	n += o

	return n, err
}

func Default() Config {
	return Config{
		CacheDir: defaultCacheDir,
		Log:      defaultLogConfig(),
		DNS:      defaultDNSConfig(),
	}
}

func (m *Config) Init() error {
	m.Log.Init()
	return m.DNS.Init(m.CacheDir)
}

func (m Config) Start() error {
	return m.DNS.Start()
}

func (m Config) Shutdown() error {
	return m.DNS.Shutdown()
}
