package bdconfig

import "time"

const (
	defaultBlockTTL       = 1 * time.Hour
	defaultUpdateInterval = 24 * time.Hour
)

type duration struct {
	time.Duration
}

func (d duration) Default(name string) interface{} {
	switch name {
	case "TTL":
		return defaultBlockTTL
	case "UpdateInterval":
		return defaultUpdateInterval
	}
	return d
}

func (d duration) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret duration
	err := ret.UnmarshalText([]byte(text))
	return ret, err
}

func (d duration) Equal(val interface{}) bool {
	if tval, ok := val.(time.Duration); ok {
		return d.Duration == tval
	}
	return false
}

func (d *duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	d.Duration = tmp
	return nil
}
