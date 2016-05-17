package bdtype

import "net"

type IP struct {
	net.IP
}

func (i IP) Default(name string) interface{} {
	return ""
}

func (i IP) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret IP
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *IP) UnmarshalText(text []byte) error {
	i.IP = net.ParseIP(string(text))
	return nil
}
