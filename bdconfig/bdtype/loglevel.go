package bdtype

import (
	"encoding"

	"jrubin.io/cliconfig"
	"jrubin.io/slog"
)

const DefaultLogLevel = LogLevel(slog.WarnLevel)

type LogLevel slog.Level

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
		return l.Level() == slog.ParseLevel(sval, DefaultLogLevel.Level())
	}
	return false
}

func (l LogLevel) Level() slog.Level {
	return slog.Level(l)
}

func (l LogLevel) String() string {
	return l.Level().String()
}

func (l LogLevel) MarshalText() ([]byte, error) {
	return l.Level().MarshalJSON()
}

func (l *LogLevel) UnmarshalText(text []byte) error {
	var t slog.Level
	if err := t.UnmarshalText(text); err != nil {
		return err
	}
	*l = LogLevel(t)
	return nil
}

func (l LogLevel) Default(name string) interface{} {
	return DefaultLogLevel.String()
}
