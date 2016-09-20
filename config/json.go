package config

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type DNSZone struct {
	Name string   `toml:"name" json:"name"`
	Addr []string `toml:"addr" json:"addr"`
}

type DNSZones []DNSZone

func (z *DNSZones) Set(value string) error {
	if err := json.Unmarshal([]byte(value), z); err != nil {
		return errors.Wrapf(err, "config.DNSZones: error unmarshaling json: %s", value)
	}
	return nil
}

func (z DNSZones) String() string {
	b, _ := json.Marshal(z)
	return string(b)
}

func (z DNSZones) Generic() cli.Generic {
	return &z
}

type StringMapStringSlice map[string][]string

func (m *StringMapStringSlice) Set(value string) error {
	if err := json.Unmarshal([]byte(value), m); err != nil {
		return errors.Wrapf(err, "config.StringMapStringSlice: error unmarshaling json: %s", value)
	}
	return nil
}

func (m StringMapStringSlice) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (m StringMapStringSlice) Generic() cli.Generic {
	return &m
}
