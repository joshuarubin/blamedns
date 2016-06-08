package config

import "time"

type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return d.Duration().String()
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(tmp)
	return nil
}

func ParseDuration(val time.Duration) Duration {
	return Duration(val)
}
