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

func (c config) Write(w io.Writer) {
	c.Log.Write(w)
	c.DNS.Write(w)
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

func (d dnsConfig) Write(w io.Writer) {
	w.Write([]byte("[dns]\n"))

	writeSS(w, "forward", d.Forward)

	for k, v := range d.Servers {
		w.Write([]byte(fmt.Sprintf("[dns.servers.%s]\n", k)))

		if len(v.Addr) > 0 {
			w.Write([]byte(fmt.Sprintf("addr = \"%s\"\n", v.Addr)))
		}

		if len(v.Net) > 0 && v.Net != "udp" {
			w.Write([]byte(fmt.Sprintf("net  = \"%s\"\n", v.Net)))
		}
	}

	d.Block.Write(w)
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
			IP:  net.IP(d.Block.IP),
			TTL: time.Duration(d.Block.TTL),
		},
	}
}

type blockConfig struct {
	IP    ip
	TTL   duration
	Hosts []string
}

func writeSS(w io.Writer, name string, ss []string) {
	if len(ss) == 0 {
		return
	}

	w.Write([]byte(fmt.Sprintf("%s = [\n", name)))

	for _, h := range ss {
		w.Write([]byte(fmt.Sprintf("  \"%s\",\n", h)))
	}

	w.Write([]byte("]\n"))
}

func (b blockConfig) Write(w io.Writer) {
	w.Write([]byte("[dns.block]\n"))

	if b.IP != nil {
		w.Write([]byte(fmt.Sprintf(
			"ip = \"%s\"\n",
			net.IP(b.IP).String(),
		)))
	}

	if b.TTL != 0 {
		w.Write([]byte(fmt.Sprintf(
			"ttl = \"%s\"\n",
			time.Duration(b.TTL).String(),
		)))
	}

	writeSS(w, "hosts", b.Hosts)
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
