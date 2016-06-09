package config

import "net"

type IP net.IP

func (i IP) IP() net.IP {
	return net.IP(i)
}

func (i IP) String() string {
	return i.IP().String()
}

func (i IP) MarshalText() ([]byte, error) {
	return i.IP().MarshalText()
}

func (i *IP) UnmarshalText(text []byte) error {
	var tmp net.IP
	err := tmp.UnmarshalText(text)
	if err != nil {
		return err
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
