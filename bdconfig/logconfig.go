package bdconfig

import (
	"github.com/Sirupsen/logrus"

	"jrubin.io/blamedns/bdconfig/bdtype"
)

type LogConfig struct {
	File  bdtype.LogFile  `toml:"file" cli:",set log filename (stderr, stdout or any file name)"`
	Level bdtype.LogLevel `toml:"level" cli:",set log level (debug, info, warning, error)"`
}

func defaultLogConfig() LogConfig {
	return LogConfig{
		File:  bdtype.DefaultLogFile(),
		Level: bdtype.DefaultLogLevel,
	}
}

func (cfg LogConfig) Init(root *Config) {
	logger := root.Logger
	logger.WithField("name", cfg.File.Name).Info("log location")
	logger.Out = cfg.File

	if cfg.File.IsFile {
		logger.Formatter = &logrus.TextFormatter{
			DisableColors: true,
		}
	}

	l := logrus.Level(cfg.Level)
	logger.Level = l
	logger.WithField("level", l).Debug("log level set")
}
