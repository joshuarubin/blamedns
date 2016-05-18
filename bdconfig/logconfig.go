package bdconfig

import (
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"

	"jrubin.io/blamedns/bdconfig/bdtype"
)

type LogConfig struct {
	File  bdtype.LogFile  `cli:",set log filename (stderr, stdout or any file name)"`
	Level bdtype.LogLevel `cli:",set log level (debug, info, warning, error)"`
}

func defaultLogConfig() LogConfig {
	return LogConfig{
		File:  bdtype.DefaultLogFile(),
		Level: bdtype.DefaultLogLevel(),
	}
}

func (cfg LogConfig) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "[log]\n")
	if err != nil {
		return n, err
	}

	var o int
	l := logrus.Level(cfg.Level).String()
	if l != "unknown" {
		o, err = fmt.Fprintf(w, "level = \"%s\"\n", l)
		n += o
		if err != nil {
			return n, err
		}
	}

	if len(cfg.File.Name) > 0 {
		o, err = fmt.Fprintf(w, "file = \"%s\"\n", cfg.File.Name)
		n += o
		if err != nil {
			return n, err
		}
	}

	return n, nil
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
