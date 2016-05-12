package main

import (
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	defaultLogLevel       = log.WarnLevel
	defaultLogLevelString = "WARN"
	defaultLogFile        = "stderr"
)

type logConfig struct {
	Level string
	File  string
}

func (cfg logConfig) Write(w io.Writer) {
	w.Write([]byte("[log]\n"))

	if len(cfg.Level) > 0 {
		w.Write([]byte(fmt.Sprintf(
			"level = \"%s\"\n",
			cfg.Level,
		)))
	}

	if len(cfg.File) > 0 {
		w.Write([]byte(fmt.Sprintf(
			"file  = \"%s\"\n",
			cfg.File,
		)))
	}
}

var logFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "log-level, l",
		EnvVar: "LOG_LEVEL",
		Value:  defaultLogLevelString,
		Usage:  "set log level (DEBUG, INFO, WARN, ERR)",
	},
	cli.StringFlag{
		Name:   "log-file, f",
		EnvVar: "LOG_FILE",
		Value:  defaultLogFile,
		Usage:  "set log filename (stderr, stdout or any file name)",
	},
}

func parseLogFileName(file string) (io.Writer, string, bool) {
	switch file {
	case "stderr", "STDERR":
		return os.Stderr, "stderr", false
	case "stdout", "STDOUT":
		return os.Stdout, "stdout", false
	default:
		f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		// file is only closed on os.Exit
		if err != nil {
			log.WithFields(log.Fields{
				"file": file,
				"err":  err,
			}).Fatal("could not open file")
		}
		return f, file, true
	}
}

func (cfg *logConfig) parseFlags(c *cli.Context) {
	level := c.String("log-level")
	setCfg(&cfg.Level, level, level == defaultLogLevelString)

	file := c.String("log-file")
	setCfg(&cfg.File, file, file == defaultLogFile)
}

func (cfg logConfig) init() {
	w, name, isFile := parseLogFileName(cfg.File)
	log.WithField("name", name).Info("log location")
	log.SetOutput(w)

	if isFile {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
		})
	}

	level := cfg.Level

	l := defaultLogLevel

	if len(level) > 0 {
		r, _ := utf8.DecodeRuneInString(level)
		r = unicode.ToLower(r)

		switch r {
		case 'e':
			l = log.ErrorLevel
		case 'i':
			l = log.InfoLevel
		case 'd':
			l = log.DebugLevel
		}
	}

	log.SetLevel(l)
	log.WithField("level", l).Debug("log level set")
}
