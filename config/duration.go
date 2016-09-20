package config

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

type Duration time.Duration

func (d Duration) Value() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return d.Value().String()
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

func (d *Duration) Set(value string) error {
	tmp, err := time.ParseDuration(value)
	if err != nil {
		return errors.Wrapf(err, "config.Duration: error parsing duration: %s", value)
	}
	*d = Duration(tmp)
	return nil
}

func (d Duration) Generic() cli.Generic {
	return &d
}
