package main

import (
	"os"
	"syscall"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
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

func parseConfigFile(c *cli.Context) {
	file := c.String("config")

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
