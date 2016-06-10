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

func (c *LogConfig) Flags(prefix string) []cli.Flag {
	return []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        flagName(prefix, "file"),
			EnvVar:      envName(prefix, "FILE"),
			Usage:       "log file name",
			Value:       c.File,
			Destination: &c.File,
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:        flagName(prefix, "level"),
			EnvVar:      envName(prefix, "LEVEL"),
			Usage:       "set log level (debug, info, warn, error)",
			Value:       c.Level,
			Destination: &c.Level,
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:        flagName(prefix, "json"),
			EnvVar:      envName(prefix, "JSON"),
			Usage:       "use json formatting when logging",
			Destination: &c.JSON,
		}),
	}
}
