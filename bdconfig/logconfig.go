package bdconfig

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
)

type LogConfig struct {
	File  logFile  `cli:",set log filename (stderr, stdout or any file name)"`
	Level logLevel `cli:",set log level (DEBUG, INFO, WARN, ERR)"`
}

func defaultLogConfig() LogConfig {
	return LogConfig{
		File:  *defaultLogFile(),
		Level: defaultLogLevel(),
	}
}

func (cfg LogConfig) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "[log]\n")
	if err != nil {
		return n, err
	}

	var o int
	l := log.Level(cfg.Level).String()
	if l != "unknown" {
		o, err = fmt.Fprintf(w, "level = \"%s\"\n", l)
		n += o
		if err != nil {
			return n, err
		}
	}

	if len(cfg.File.Name) > 0 {
		o, err = fmt.Fprintf(w, "file  = \"%s\"\n", cfg.File.Name)
		n += o
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (cfg LogConfig) Init() {
	log.WithField("name", cfg.File.Name).Info("log location")
	log.SetOutput(cfg.File)

	if cfg.File.IsFile {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
		})
	}

	l := log.Level(cfg.Level)
	log.SetLevel(l)
	log.WithField("level", l).Debug("log level set")
}
