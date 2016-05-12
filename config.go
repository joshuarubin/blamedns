package main

import (
	"fmt"
	"io"
	"net"
	"strings"
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

func (d dnsConfig) Write(w io.Writer) {
	w.Write([]byte("[dns]\n"))

	if len(d.Forward) > 0 {
		w.Write([]byte(fmt.Sprintf(
			"forward = [\n"+
				"  \"%s\"\n"+
				"]\n",
			strings.Join(d.Forward, "\",\n  \""),
		)))
	}

	for k, v := range d.Servers {
		if len(k) == 0 || len(v.Addr) == 0 || len(v.Net) == 0 {
			continue
		}

		w.Write([]byte(fmt.Sprintf(
			"[dns.servers.%s]\n"+
				"addr = \"%s\"\n"+
				"net  = \"%s\"\n",
			k,
			v.Addr,
			v.Net,
		)))
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

	if len(b.Hosts) > 0 {
		w.Write([]byte(fmt.Sprintf(
			"hosts = [\n"+
				"  \"%s\"\n"+
				"]\n",
			strings.Join(b.Hosts, "\",\n  \""),
		)))
	}
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
