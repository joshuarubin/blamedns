package main // import "jrubin.io/blamedns"

import (
	"os"
	"os/signal"
	"syscall"

	"jrubin.io/blamedns/bdconfig"
	"jrubin.io/cliconfig"

	"github.com/codegangsta/cli"
)

var (
	name, version string

	cfg = bdconfig.New(name, version)
	cc  = cliconfig.New(bdconfig.Default())
	app = cli.NewApp()

	logger = cfg.Logger
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
		logger.Fatal(err)
	}
}

func setup(c *cli.Context) error {
	parseConfigFile(c)
	return cc.Parse(c, cfg)
}

func run(c *cli.Context) error {
	if err := cfg.Init(); err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- cfg.Start()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case sig := <-sigs:
		logger.WithField("signal", sig).Debug("received signal")
	}

	return cfg.Shutdown()
}

func writeConfig(c *cli.Context) error {
	_, err := cfg.Write(os.Stdout)
	return err
}
