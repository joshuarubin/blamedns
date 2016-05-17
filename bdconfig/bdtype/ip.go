package bdtype

import "net"

type IP net.IP

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

func (i *IP) UnmarshalText(text []byte) error {
	*i = IP(net.ParseIP(string(text)))
	return nil
}
