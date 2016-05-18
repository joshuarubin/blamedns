package bdtype

import (
	"encoding"
	"unicode"
	"unicode/utf8"

	"jrubin.io/cliconfig"

	"github.com/Sirupsen/logrus"
)

const DefaultLogLevel = LogLevel(logrus.WarnLevel)

type LogLevel logrus.Level

var (
	tmpLvl                          = LogLevel(0)
	_      cliconfig.CustomType     = tmpLvl
	_      encoding.TextMarshaler   = tmpLvl
	_      encoding.TextUnmarshaler = &tmpLvl
)

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

func (l LogLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *LogLevel) UnmarshalText(text []byte) error {
	*l = parseLogLevel(string(text))
	return nil
}

func (l LogLevel) Default(name string) interface{} {
	return DefaultLogLevel.String()
}

func parseLogLevel(level string) LogLevel {
	ret := DefaultLogLevel

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

	return ret
}
