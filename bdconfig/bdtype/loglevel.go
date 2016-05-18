package bdtype

import (
	"unicode"
	"unicode/utf8"

	"jrubin.io/cliconfig"

	"github.com/Sirupsen/logrus"
)

const defaultLogLevelValue = LogLevel(logrus.WarnLevel)

type LogLevel logrus.Level

var _ cliconfig.CustomType = LogLevel(0)

func DefaultLogLevel() LogLevel {
	return LogLevel(defaultLogLevelValue)
}

func (l LogLevel) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret LogLevel
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (l LogLevel) Equal(val interface{}) bool {
	if sval, ok := val.(string); ok {
		return l == parseLogLevel(sval)
	}
	return false
}

func (l LogLevel) Level() logrus.Level {
	return logrus.Level(l)
}

func (l LogLevel) String() string {
	return l.Level().String()
}

func (l *LogLevel) UnmarshalText(text []byte) error {
	*l = parseLogLevel(string(text))
	return nil
}

func (l LogLevel) Default(name string) interface{} {
	return defaultLogLevelValue.String()
}

func parseLogLevel(level string) LogLevel {
	ret := defaultLogLevelValue

	if len(level) > 0 {
		r, _ := utf8.DecodeRuneInString(level)
		r = unicode.ToLower(r)

		switch r {
		case 'e':
			return LogLevel(logrus.ErrorLevel)
		case 'i':
			return LogLevel(logrus.InfoLevel)
		case 'd':
			return LogLevel(logrus.DebugLevel)
		}
	}

	return LogLevel(ret)
}
