package config

import (
	"net"

	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

type IP net.IP

func (i IP) Value() net.IP {
	return net.IP(i)
}

func (i IP) String() string {
	return i.Value().String()
}

func (i IP) MarshalText() ([]byte, error) {
	data, err := i.Value().MarshalText()
	if err != nil {
		return nil, errors.Wrapf(err, "config.IP: error marshaling text: %s", i)
	}
	return data, nil
}

func (i *IP) UnmarshalText(text []byte) error {
	var tmp net.IP
	err := tmp.UnmarshalText(text)
	if err != nil {
		return errors.Wrapf(err, "config.IP: error unmarshaling json: %s", string(text))
	}
	*i = IP(tmp)
	return nil
}

func ParseIP(value string) IP {
	return IP(net.ParseIP(value))
}

func (i *IP) Set(value string) error {
	return i.UnmarshalText([]byte(value))
}

func (i IP) Generic() cli.Generic {
	return &i
}
