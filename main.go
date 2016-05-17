package main // import "jrubin.io/blamedns"

import (
	"os"
	"os/signal"
	"syscall"

	"jrubin.io/blamedns/bdconfig"
	"jrubin.io/cliconfig"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	defaultConfigFile = "/etc/blamedns/config.toml"
)

var (
	name, version string

	cfg bdconfig.Config
	app = cli.NewApp()
	cc  = cliconfig.New(bdconfig.Default())
)

func init() {
	app.Name = name
	app.Version = version
	app.Usage = "" // TODO(jrubin)
	app.Authors = []cli.Author{{
		Name:  "Joshua Rubin",
		Email: "joshua@rubixconsulting.com",
	}}
	app.Before = setup
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			EnvVar: "BLAMEDNS_CONFIG",
			Value:  defaultConfigFile,
			Usage:  "config file",
		},
	}
	app.Flags = append(app.Flags, cc.Flags()...)

	app.Commands = []cli.Command{{
		Name:   "config",
		Usage:  "write the config to stdout",
		Action: writeConfig,
	}}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func parseConfigFile(file string) {
	if len(file) == 0 {
		return
	}

	md, err := toml.DecodeFile(file, &cfg)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err != syscall.ENOENT {
			log.WithError(err).Warn("error reading config file")
		}
		return
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		log.Warnf("undecoded keys: %q", undecoded)
	}
}

func setup(c *cli.Context) error {
	parseConfigFile(c.String("config"))

	if err := cc.Parse(c, &cfg); err != nil {
		return err
	}

	return cfg.Init()
}

func run(c *cli.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- cfg.DNS.Server.ListenAndServe()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case sig := <-sigs:
		log.WithField("signal", sig).Debug("received signal")
	}

	return cfg.DNS.Server.Shutdown()
}

func writeConfig(c *cli.Context) error {
	_, err := cfg.Write(os.Stdout)
	return err
}
