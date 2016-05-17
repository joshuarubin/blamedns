package bdconfig

import "net"

type ip struct {
	net.IP
}

func (i ip) Default(name string) interface{} {
	return ""
}

func (i ip) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret ip
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (i *ip) UnmarshalText(text []byte) error {
	i.IP = net.ParseIP(string(text))
	return nil
}
