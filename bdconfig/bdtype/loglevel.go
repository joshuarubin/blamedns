package bdconfig

import (
	"unicode"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
)

const defaultLogLevelValue = logLevel(log.WarnLevel)

type logLevel log.Level

func defaultLogLevel() logLevel {
	return logLevel(defaultLogLevelValue)
}

func (l logLevel) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret logLevel
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (l logLevel) Equal(val interface{}) bool {
	if sval, ok := val.(string); ok {
		return l == parseLogLevel(sval)
	}
	return false
}

func (l *logLevel) UnmarshalText(text []byte) error {
	*l = parseLogLevel(string(text))
	return nil
}

func (l logLevel) Default(name string) interface{} {
	return log.Level(defaultLogLevelValue).String()
}

func parseLogLevel(level string) logLevel {
	ret := defaultLogLevelValue

	if len(level) > 0 {
		r, _ := utf8.DecodeRuneInString(level)
		r = unicode.ToLower(r)

		switch r {
		case 'e':
			return logLevel(log.ErrorLevel)
		case 'i':
			return logLevel(log.InfoLevel)
		case 'd':
			return logLevel(log.DebugLevel)
		}
	}

	return logLevel(ret)
}
