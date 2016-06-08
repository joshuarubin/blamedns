package main // import "jrubin.io/blamedns"

import (
	"os"
	"os/signal"
	"syscall"

	"jrubin.io/blamedns/config"

	"github.com/joshuarubin/cli"
)

var (
	name, version string

	cfg = config.New(name, version)
	app = cli.NewApp()
)

func init() {
	app.Name = name
	app.Version = version
	app.Usage = "a simple blocking and caching recursive dns server"
	app.Authors = []cli.Author{{
		Name:  "Joshua Rubin",
		Email: "joshua@rubixconsulting.com",
	}}
	app.Before = configSetup
	app.Action = run
	app.Flags = append(configFileFlags, config.Flags(cfg)...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		cfg.Logger.WithError(err).Fatal("application error")
	}
}

func run(c *cli.Context) error {
	if err := cfg.Init(); err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() { errCh <- cfg.Start() }()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case sig := <-sigs:
			cfg.Logger.WithField("signal", sig).Debug("received signal")
			switch sig {
			case syscall.SIGUSR1:
				cfg.SIGUSR1()
			default:
				cfg.Shutdown()
				return nil
			}
		}
	}
}
