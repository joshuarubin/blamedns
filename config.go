package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"jrubin.io/blamedns/dnsserver"

	"github.com/miekg/dns"
)

type config struct {
	Log logConfig
	DNS dnsConfig
}

func (c config) Write(w io.Writer) (int, error) {
	n, err := c.Log.Write(w)
	if err != nil {
		return n, err
	}

	o, err := c.DNS.Write(w)
	return n + o, err
}

type dnsConfig struct {
	Forward []string
	Servers map[string]*dns.Server
	Block   blockConfig
}

func (d *dnsConfig) SetDefaults() {
	for _, server := range d.Servers {
		if len(server.Addr) > 0 && len(server.Net) == 0 {
			server.Net = "udp"
		}
	}
}

func (d dnsConfig) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "[dns]\n")
	if err != nil {
		return n, err
	}

	o, err := writeSS(w, "forward", d.Forward)
	n += o
	if err != nil {
		return n, err
	}

	for k, v := range d.Servers {
		o, err = fmt.Fprintf(w, "[dns.servers.%s]\n", k)
		n += o
		if err != nil {
			return n, err
		}

		if len(v.Addr) > 0 {
			o, err = fmt.Fprintf(w, "addr = \"%s\"\n", v.Addr)
			n += o
			if err != nil {
				return n, err
			}
		}

		if len(v.Net) > 0 && v.Net != "udp" {
			o, err = fmt.Fprintf(w, "net  = \"%s\"\n", v.Net)
			n += o
			if err != nil {
				return n, err
			}
		}
	}

	o, err = d.Block.Write(w)
	return n + o, err
}

func (d dnsConfig) ServersSlice() []*dns.Server {
	ret := make([]*dns.Server, len(d.Servers))
	i := 0
	for _, s := range d.Servers {
		ret[i] = s
		i++
	}
	return ret
}

func (d dnsConfig) Server() *dnsserver.DNSServer {
	return &dnsserver.DNSServer{
		Servers: d.ServersSlice(),
		Forward: d.Forward,
		Block: dnsserver.Block{
			IPv4: net.IP(d.Block.IPv4),
			IPv6: net.IP(d.Block.IPv6),
			Host: d.Block.Host,
			TTL:  time.Duration(d.Block.TTL),
		},
	}
}

type blockConfig struct {
	IPv4, IPv6 ip
	Host       string
	TTL        duration
	Hosts      []string
}

func writeSS(w io.Writer, name string, ss []string) (int, error) {
	if len(ss) == 0 {
		return 0, nil
	}

	n, err := fmt.Fprintf(w, "%s = [\n", name)
	if err != nil {
		return n, err
	}

	for _, h := range ss {
		var o int
		o, err = fmt.Fprintf(w, "  \"%s\",\n", h)
		n += o
		if err != nil {
			return n, err
		}
	}

	o, err := fmt.Fprintf(w, "]\n")
	return n + o, err
}

func (b blockConfig) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "[dns.block]\n")
	if err != nil {
		return n, err
	}

	var o int
	if b.IPv4 != nil {
		o, err = fmt.Fprintf(w, "ipv4 = \"%s\"\n", net.IP(b.IPv4).String())
		n += o
		if err != nil {
			return n, err
		}
	}

	if b.IPv6 != nil {
		o, err = fmt.Fprintf(w, "ipv6 = \"%s\"\n", net.IP(b.IPv6).String())
		n += o
		if err != nil {
			return n, err
		}
	}

	if len(b.Host) > 0 {
		o, err = fmt.Fprintf(w, "host = \"%s\"\n", b.Host)
		n += o
		if err != nil {
			return n, err
		}
	}

	if b.TTL != 0 {
		o, err = fmt.Fprintf(w, "ttl = \"%s\"\n", time.Duration(b.TTL).String())
		n += o
		if err != nil {
			return n, err
		}
	}

	o, err = writeSS(w, "hosts", b.Hosts)
	return n + o, err
}

type duration time.Duration

func (d *duration) UnmarshalText(text []byte) error {
	t, err := time.ParseDuration(string(text))
	*d = duration(t)
	return err
}

type ip net.IP

func (i *ip) UnmarshalText(text []byte) error {
	t := net.ParseIP(string(text))
	*i = ip(t)
	return nil
}
