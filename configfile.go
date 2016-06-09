package main

import (
	"io"
	"os"

	"jrubin.io/blamedns/config"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"
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

func setupWriteConfig(c *cli.Context) error {
	var err error
	configOut, err = config.ParseFileFlag(c.String("config-out"))
	return err
}

func writeConfig(c *cli.Context) error {
	enc := toml.NewEncoder(configOut)
	enc.Indent = ""
	return enc.Encode(cfg)
}
