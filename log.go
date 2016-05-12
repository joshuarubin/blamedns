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
	Level logLevel
	File  logFile
}

type logLevel log.Level

func (l *logLevel) UnmarshalText(text []byte) error {
	*l = logLevel(parseLogLevel(string(text)))
	return nil
}

type logFile struct {
	io.Writer
	Name   string
	IsFile bool
}

func (l *logFile) UnmarshalText(text []byte) error {
	t, err := parseLogFileName(string(text))
	if err != nil {
		return err
	}
	*l = *t
	return nil
}

func (cfg logConfig) Write(w io.Writer) {
	w.Write([]byte("[log]\n"))

	l := log.Level(cfg.Level).String()
	if l != "unknown" {
		w.Write([]byte(fmt.Sprintf("level = \"%s\"\n", l)))
	}

	if len(cfg.File.Name) > 0 {
		w.Write([]byte(fmt.Sprintf("file  = \"%s\"\n", cfg.File.Name)))
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

func parseLogFileName(file string) (*logFile, error) {
	switch file {
	case "stderr", "STDERR":
		return &logFile{
			Writer: os.Stderr,
			Name:   "stderr",
			IsFile: false,
		}, nil
	case "stdout", "STDOUT":
		return &logFile{
			Writer: os.Stdout,
			Name:   "stdout",
			IsFile: false,
		}, nil
	default:
		f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		// file is only closed on os.Exit
		if err != nil {
			return nil, err
		}

		return &logFile{
			Writer: f,
			Name:   file,
			IsFile: true,
		}, nil
	}
}

func (cfg *logConfig) parseFlags(c *cli.Context) error {
	level := parseLogLevel(c.String("log-level"))
	setCfg(&cfg.Level, logLevel(level), level == defaultLogLevel)

	file, err := parseLogFileName(c.String("log-file"))
	if err != nil {
		return err
	}

	setCfg(&cfg.File, *file, file.Name == defaultLogFile)

	return nil
}

func parseLogLevel(level string) log.Level {
	ret := defaultLogLevel

	if len(level) > 0 {
		r, _ := utf8.DecodeRuneInString(level)
		r = unicode.ToLower(r)

		switch r {
		case 'e':
			return log.ErrorLevel
		case 'i':
			return log.InfoLevel
		case 'd':
			return log.DebugLevel
		}
	}

	return ret
}

func (cfg logConfig) init() {
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
