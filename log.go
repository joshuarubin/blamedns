package main

import (
	"unicode"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	defaultLogLevel       = log.WarnLevel
	defaultLogLevelString = "WARN"
)

var logFlag = cli.StringFlag{
	Name:   "log-level, l",
	EnvVar: "LOG_LEVEL",
	Value:  defaultLogLevelString,
	Usage:  "set log level (DEBUG, INFO, WARN, ERR)",
}

func setLogLevel(c *cli.Context) {
	level := c.String("log-level")

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
