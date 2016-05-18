package bdconfig

import (
	"fmt"
	"io"
	"net/url"

	"github.com/miekg/dns"

	"jrubin.io/blamedns/bdconfig/stringslice"
	"jrubin.io/blamedns/dnsserver"
)

type DNSConfig struct {
	Forward []string `cli:",dns server(s) to forward requests to"`
	Listen  []string `cli:",url(s) to listen for dns requests on"`
	Block   BlockConfig
	Server  *dnsserver.DNSServer `cli:"-"`
}

func defaultDNSConfig() DNSConfig {
	return DNSConfig{
		Forward: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
		Listen: []string{
			"udp://[::]:53",
			"tcp://[::]:53",
		},
		Block: defaultBlockConfig(),
	}
}

func (cfg DNSConfig) Write(w io.Writer) (int, error) {
	n, err := fmt.Fprintf(w, "[dns]\n")
	if err != nil {
		return n, err
	}

	o, err := stringslice.Write("forward", cfg.Forward, w)
	n += o
	if err != nil {
		return n, err
	}

	o, err = stringslice.Write("listen", cfg.Listen, w)
	n += o
	if err != nil {
		return n, err
	}

	o, err = cfg.Block.Write(w)
	n += o

	return n, err
}

func parseDNSServer(val string) (*dns.Server, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}

	return &dns.Server{
		Addr: u.Host,
		Net:  u.Scheme,
	}, nil
}

func (cfg *DNSConfig) Init(cacheDir string) error {
	servers := make([]*dns.Server, len(cfg.Listen))

	for i, listen := range cfg.Listen {
		server, err := parseDNSServer(listen)
		if err != nil {
			return err
		}
		servers[i] = server
	}

	if err := cfg.Block.Init(cacheDir); err != nil {
		return err
	}

	cfg.Server = &dnsserver.DNSServer{
		Forward: cfg.Forward,
		Servers: servers,
		Block:   cfg.Block.Block(),
	}

	return nil
}

func (cfg DNSConfig) Start() error {
	if err := cfg.Block.Start(); err != nil {
		return err
	}

	return cfg.Server.ListenAndServe()
}

func (cfg DNSConfig) Shutdown() error {
	if err := cfg.Server.Shutdown(); err != nil {
		return err
	}

	return cfg.Block.Shutdown()
}
