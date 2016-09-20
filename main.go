package main // import "jrubin.io/blamedns"

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"jrubin.io/blamedns/config"
	"jrubin.io/blamedns/context"
	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"

	"github.com/BurntSushi/toml"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

const defaultConfigFile = "/etc/blamedns/config.toml"

var (
	version   string
	configOut io.Writer

	logger = text.Logger(slog.WarnLevel)
	cfg    = config.New()
	app    = cli.NewApp()
)

func init() {
	app.Version = version
	app.Usage = "a simple blocking and caching recursive dns server"
	app.Authors = []cli.Author{{
		Name:  "Joshua Rubin",
		Email: "joshua@rubixconsulting.com",
	}}
	app.Flags = append(
		[]cli.Flag{
			cli.StringFlag{
				Name:   "config",
				EnvVar: "BLAMEDNS_CONFIG",
				Value:  defaultConfigFile,
				Usage:  "config file",
			},
		},
		cfg.Flags()...,
	)
	app.Before = altsrc.InitInputSourceWithContext(
		app.Flags,
		config.TOMLSource(defaultConfigFile, "config"),
	)
	app.Action = run
	app.Commands = append(app.Commands, cli.Command{
		Name:   "config",
		Usage:  "write the config to stdout",
		Before: setupWriteConfig,
		Action: writeConfig,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "config-out",
				EnvVar: "CONFIG_OUT",
				Usage:  "file to write config to",
				Value:  os.Stdout.Name(),
			},
		},
	})
}

func main() {
	if err := app.Run(os.Args); err != nil {
		logger.WithError(err).Fatal("application error")
	}
}

func run(c *cli.Context) error {
	ctx, err := context.New(app.Name, app.Version, cfg)
	if err != nil {
		return err
	}
	logger = ctx.Log.Logger

	errCh := make(chan error, 1)
	go func() { errCh <- ctx.Start() }()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case sig := <-sigs:
			logger.WithField("signal", sig).Debug("received signal")
			switch sig {
			case syscall.SIGUSR1:
				ctx.SIGUSR1()
			default:
				ctx.Shutdown()
				return nil
			}
		}
	}
}

func setupWriteConfig(c *cli.Context) error {
	var err error
	configOut, err = context.ParseFileFlag(c.String("config-out"))
	return err
}

func writeConfig(c *cli.Context) error {
	enc := toml.NewEncoder(configOut)
	enc.Indent = ""
	return enc.Encode(cfg)
}
