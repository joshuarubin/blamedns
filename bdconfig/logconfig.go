package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/slog/handlers/json"
	"jrubin.io/slog/handlers/text"
)

type LogConfig struct {
	File  bdtype.LogFile  `toml:"file" cli:",set log filename (stderr, stdout or any file name [if a file, the output will be in json])"`
	Level bdtype.LogLevel `toml:"level" cli:",set log level (debug, info, warning, error)"`
	JSON  bool            `toml:"json" cli:",log as json data even to stderr and stdout"`
}

var defaultLogConfig = LogConfig{
	File:  bdtype.DefaultLogFile(),
	Level: bdtype.DefaultLogLevel,
}

func (cfg LogConfig) Init(root *Config) {
	logger := root.Logger
	if cfg.File.File() != nil || cfg.JSON {
		logger.RegisterHandler(cfg.Level.Level(), json.New(cfg.File.Writer))
	} else {
		logger.RegisterHandler(cfg.Level.Level(), text.New(cfg.File.Writer))
	}
	logger.WithField("level", cfg.Level).Debug("log level set")
}

func (cfg LogConfig) Shutdown() {
	if f := cfg.File.File(); f != nil {
		_ = f.Close()
	}
}
