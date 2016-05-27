package bdtype

import (
	"encoding"
	"fmt"
	"net"

	"jrubin.io/cliconfig"
)

type IP net.IP

var (
	_ cliconfig.CustomType     = IP{}
	_ encoding.TextMarshaler   = IP{}
	_ encoding.TextUnmarshaler = &IP{}
)

func (i IP) Default(name string) interface{} {
	switch name {
	case "dns-block-ipv4":
		return "127.0.0.1"
	case "dns-block-ipv6":
		return "::1"
	}
	panic(fmt.Errorf("bdtype.IP.Default unknown name: %s", name))
}

func (i IP) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret IP
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i IP) Equal(val interface{}) bool {
	if sval, ok := val.(string); ok {
		return i.String() == sval
	}
	return false
}

func (i IP) IP() net.IP {
	return net.IP(i)
}

func (i IP) String() string {
	return i.IP().String()
}

func (i IP) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *IP) UnmarshalText(text []byte) error {
	*i = IP(net.ParseIP(string(text)))
	return nil
}
