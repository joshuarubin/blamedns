package config

import (
	"os"
	"strconv"
	"syscall"

	"jrubin.io/inputsource"
	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"

	"github.com/BurntSushi/toml"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

func TOMLSource(dflt, key string) func(*cli.Context) (altsrc.InputSourceContext, error) {
	return func(c *cli.Context) (altsrc.InputSourceContext, error) {
		file := c.String(key)
		logger := text.Logger(slog.WarnLevel).WithField("config_file", file)
		ret := New()

		md, err := toml.DecodeFile(file, ret)
		if err != nil {
			perr, ok := err.(*os.PathError)
			if !ok || perr.Err != syscall.ENOENT || file != dflt {
				// ignore file not found error if the config file was the default one
				logger.WithError(err).Warn("error loading toml config")
			}
			return inputsource.New(ret), nil
		}

		if undecoded := md.Undecoded(); len(undecoded) > 0 {
			var ctxLog slog.Interface = logger
			for i, k := range undecoded {
				ctxLog = ctxLog.WithField("key_"+strconv.Itoa(i), k.String())
			}
			ctxLog.Warn("undecoded keys")
		}

		return inputsource.New(ret), nil
	}
}
