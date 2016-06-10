package config

import (
	"os"

	"jrubin.io/slog"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var defaultLogLevel = slog.WarnLevel

type LogConfig struct {
	File  string `toml:"file"`
	Level string `toml:"level"`
	JSON  bool   `toml:"json"`
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		File:  os.Stderr.Name(),
		Level: defaultLogLevel.String(),
	}
}

func (c *LogConfig) Flags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        "log-file",
			EnvVar:      "LOG_FILE",
			Usage:       "log file name",
			Value:       c.File,
			Destination: &c.File,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        "log-level",
			EnvVar:      "LOG_LEVEL",
			Usage:       "set log level (debug, info, warn, error)",
			Value:       c.Level,
			Destination: &c.Level,
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        "log-json",
			EnvVar:      "LOG_JSON",
			Usage:       "use json formatting when logging",
			Destination: &c.JSON,
		}),
	}
}

func (c *LogConfig) Get(name string) (interface{}, bool) {
	switch name {
	case "log-file":
		return c.File, true
	case "log-level":
		return c.Level, true
	case "log-json":
		return c.JSON, true
	}

	return nil, false
}
