package bdconfig

import (
	"jrubin.io/blamedns/bdconfig/bdtype"
	"jrubin.io/slog/handlers/auto"
)

type LogConfig struct {
	File  bdtype.LogFile  `toml:"file" cli:",set log filename (stderr, stdout or any file name)"`
	Level bdtype.LogLevel `toml:"level" cli:",set log level (debug, info, warning, error)"`
}

var defaultLogConfig = LogConfig{
	File:  bdtype.DefaultLogFile(),
	Level: bdtype.DefaultLogLevel,
}

func (cfg LogConfig) Init(root *Config) {
	logger := root.Logger
	logger.RegisterHandler(cfg.Level.Level(), auto.New(cfg.File.Writer))
	logger.WithField("level", cfg.Level).Debug("log level set")
}

func (cfg LogConfig) Shutdown() {
	if f := cfg.File.File(); f != nil {
		_ = f.Close()
	}
}
