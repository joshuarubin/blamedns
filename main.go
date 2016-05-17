package main // import "jrubin.io/blamedns"

import (
	"os"
	"os/signal"
	"syscall"

	"jrubin.io/blamedns/bdconfig"
	"jrubin.io/cliconfig"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	name, version string

	cfg bdconfig.Config
	cc  = cliconfig.New(bdconfig.Default())
	app = cli.NewApp()
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
	app.Flags = append(configFileFlags, cc.Flags()...)
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

func setup(c *cli.Context) error {
	parseConfigFile(c)
	return cc.Parse(c, &cfg)
}

func run(c *cli.Context) error {
	if err := cfg.Init(); err != nil {
		return err
	}

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
