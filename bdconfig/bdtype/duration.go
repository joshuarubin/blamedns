package bdtype

import "time"

const (
	defaultBlockTTL       = 1 * time.Hour
	defaultUpdateInterval = 24 * time.Hour
)

type Duration struct {
	time.Duration
}

func (d Duration) Default(name string) interface{} {
	switch name {
	case "TTL":
		return defaultBlockTTL
	case "UpdateInterval":
		return defaultUpdateInterval
	}
	return d
}

func (d Duration) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret Duration
	err := ret.UnmarshalText([]byte(text))
	return ret, err
}

func (d Duration) Equal(val interface{}) bool {
	if tval, ok := val.(time.Duration); ok {
		return d.Duration == tval
	}
	return false
}

func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	d.Duration = tmp
	return nil
}
