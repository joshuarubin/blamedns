package config

import (
	"os"
	"strings"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/json"
	"jrubin.io/slog/handlers/text"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var defaultLogLevel = slog.WarnLevel

type LogConfig struct {
	File  string `toml:"file"`
	Level string `toml:"level"`
	JSON  bool   `toml:"json"`
	level slog.Level
	file  *os.File
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

func ParseFileFlag(name string) (*os.File, error) {
	lname := strings.ToLower(name)

	if lname == "stdout" || name == os.Stdout.Name() {
		return os.Stdout, nil
	}

	if lname == "stderr" || name == os.Stderr.Name() {
		return os.Stderr, nil
	}

	return os.Create(name)
}

func (c *LogConfig) Init(root *Config) error {
	logger := root.Logger
	c.level = slog.ParseLevel(c.Level, defaultLogLevel)

	var err error
	if c.file, err = ParseFileFlag(c.File); err != nil {
		return err
	}

	if c.JSON || (c.file != os.Stdout && c.file != os.Stderr) {
		logger.RegisterHandler(c.level, json.New(c.file))
	} else {
		logger.RegisterHandler(c.level, text.New(c.file))
	}

	logger.WithField("level", c.Level).Debug("log level set")
	return nil
}

func (c LogConfig) Shutdown() {
	if f := c.file; f != nil {
		_ = f.Close()
	}
}
