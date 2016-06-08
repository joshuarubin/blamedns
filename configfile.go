package main

import (
	"io"
	"os"

	"jrubin.io/blamedns/config"

	"github.com/BurntSushi/toml"
	"github.com/joshuarubin/cli"
	"github.com/joshuarubin/cli/altsrc"
)

const defaultConfigFile = "/etc/blamedns/config.toml"

var (
	configFileFlags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			EnvVar: "BLAMEDNS_CONFIG",
			Value:  defaultConfigFile,
			Usage:  "config file",
		},
	}
	configOut io.Writer
)

func init() {
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

func configSetup(c *cli.Context) error {
	if err := altsrc.InitInputSourceWithContext(app.Flags, config.InputSource(defaultConfigFile, "config"))(c); err != nil {
		return err
	}

	return cfg.Parse(c)
}

func setupWriteConfig(c *cli.Context) error {
	var err error
	configOut, err = os.Create(c.String("config-out"))
	return err
}

func writeConfig(c *cli.Context) error {
	enc := toml.NewEncoder(configOut)
	enc.Indent = ""
	return enc.Encode(cfg)
}
