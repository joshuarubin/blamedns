package config

import (
	"encoding/json"

	"github.com/urfave/cli"
)

type DNSZone struct {
	Name string   `toml:"name" json:"name"`
	Addr []string `toml:"addr" json:"addr"`
}

type DNSZones []DNSZone

func (z *DNSZones) Set(value string) error {
	return json.Unmarshal([]byte(value), z)
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
	return json.Unmarshal([]byte(value), m)
}

func (m StringMapStringSlice) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (m StringMapStringSlice) Generic() cli.Generic {
	return &m
}
