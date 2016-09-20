package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

type Config struct {
	CacheDir        string     `toml:"cache_dir"`
	ListenPixelserv string     `toml:"listen_pixelserv"`
	ListenAPIServer string     `toml:"listen_apiserver"`
	Log             *LogConfig `toml:"log"`
	DL              *DLConfig  `toml:"dl"`
	DNS             *DNSConfig `toml:"dns"`
}

func defaultCacheDir(name string) string {
	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("%s/Library/Caches/%s", os.Getenv("HOME"), name)
	}

	return fmt.Sprintf("%s/.cache/%s", os.Getenv("HOME"), name)
}

func New() *Config {
	return &Config{
		CacheDir: defaultCacheDir("blamedns"),
		Log:      NewLogConfig(),
		DL:       NewDLConfig(),
		DNS:      NewDNSConfig(),
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
			Value:       c.ListenPixelserv,
			Destination: &c.ListenPixelserv,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        "listen-apiserver",
			EnvVar:      "LISTEN_APISERVER",
			Usage:       "address to run the api server on",
			Value:       c.ListenAPIServer,
			Destination: &c.ListenAPIServer,
		}),
	}

	ret = append(ret, c.Log.Flags("log")...)
	ret = append(ret, c.DL.Flags("dl")...)
	ret = append(ret, c.DNS.Flags("dns")...)

	return ret
}

func flagName(prefix, name string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(prefix), strings.ToLower(name))
}

func envName(prefix, name string) string {
	return fmt.Sprintf("%s_%s", strings.ToUpper(prefix), strings.ToUpper(name))
}
