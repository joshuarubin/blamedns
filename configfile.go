package main

import (
	"io"
	"os"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
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
				EnvVar: "BLAMEDNS_CONFIG_OUT",
				Usage:  "file to write config to",
				Value:  "stdout",
			},
		},
	})
}

func parseConfigFile(c *cli.Context) {
	file := c.String("config")

	if len(file) == 0 {
		return
	}

	md, err := toml.DecodeFile(file, &cfg)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || perr.Err != syscall.ENOENT {
			logger.WithError(err).Warn("error reading config file")
		}
		return
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		logger.Warnf("undecoded keys: %q", undecoded)
	}
}

func setupWriteConfig(c *cli.Context) error {
	switch file := c.String("config-out"); file {
	case "stdout", "STDOUT":
		configOut = os.Stdout
	case "stderr", "STDERR":
		configOut = os.Stderr
	default:
		var err error
		if configOut, err = os.Create(file); err != nil {
			return err
		}
	}
	return nil
}

func writeConfig(c *cli.Context) error {
	return toml.NewEncoder(configOut).Encode(cfg)
}
