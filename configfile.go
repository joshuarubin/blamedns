package main

import (
	"os"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
)

const defaultConfigFile = "/etc/blamedns/config.toml"

var configFileFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "config",
		EnvVar: "BLAMEDNS_CONFIG",
		Value:  defaultConfigFile,
		Usage:  "config file",
	},
}

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "config",
		Usage:  "write the config to stdout",
		Action: writeConfig,
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

func writeConfig(c *cli.Context) error {
	return toml.NewEncoder(os.Stdout).Encode(cfg)
}
