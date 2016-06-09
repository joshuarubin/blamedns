package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

func InputSource(dflt, key string) func(*cli.Context) (altsrc.InputSourceContext, error) {
	return func(c *cli.Context) (altsrc.InputSourceContext, error) {
		file := c.String(key)
		logger := text.Logger(slog.WarnLevel).WithField("config_file", file)
		ret := New("", "")

		md, err := toml.DecodeFile(file, ret)
		if err != nil {
			perr, ok := err.(*os.PathError)
			if ok && perr.Err == syscall.ENOENT && file == dflt {
				// ignore file not found error if the config file was the default one
				return ret, nil
			}

			logger.WithError(err).Warn("error loading toml config")
			return ret, nil
		}

		if undecoded := md.Undecoded(); len(undecoded) > 0 {
			var ctxLog slog.Interface = logger
			for i, k := range undecoded {
				ctxLog = ctxLog.WithField("key_"+strconv.Itoa(i), k.String())
			}
			ctxLog.Warn("undecoded keys")
		}

		return ret, nil
	}
}

func (c *Config) Int(name string) (int, error) {
	val, ok := c.Get(name)
	if !ok {
		return 0, nil
	}

	if ret, ok := val.(int); ok {
		return ret, nil
	}

	return 0, fmt.Errorf("could not convert %T{%v} to int for %s", val, val, name)
}

func (c *Config) Duration(name string) (time.Duration, error) {
	val, ok := c.Get(name)
	if !ok {
		return 0, nil
	}

	if ret, ok := val.(time.Duration); ok {
		return ret, nil
	}

	if ret, ok := val.(Duration); ok {
		return ret.Duration(), nil
	}

	return 0, fmt.Errorf("could not convert %T{%v} to time.Duration for %s", val, val, name)
}

func (c *Config) Float64(name string) (float64, error) {
	val, ok := c.Get(name)
	if !ok {
		return 0, nil
	}

	if ret, ok := val.(float64); ok {
		return ret, nil
	}

	return 0, fmt.Errorf("could not convert %T{%v} to float64 for %s", val, val, name)
}

func (c *Config) String(name string) (string, error) {
	val, ok := c.Get(name)
	if !ok {
		return "", nil
	}

	if ret, ok := val.(string); ok {
		return ret, nil
	}

	if ret, ok := val.(net.IP); ok {
		if ret == nil {
			return "", nil
		}

		return ret.String(), nil
	}

	return "", fmt.Errorf("could not convert %T{%v} to string for %s", val, val, name)
}

func (c *Config) StringSlice(name string) ([]string, error) {
	val, ok := c.Get(name)
	if !ok {
		return nil, nil
	}

	if ret, ok := val.([]string); ok {
		return ret, nil
	}

	if ret, ok := val.(cli.StringSlice); ok {
		return ret, nil
	}

	return nil, fmt.Errorf("could not convert %T{%v} to []string for %s", val, val, name)
}

func (c *Config) IntSlice(name string) ([]int, error) {
	val, ok := c.Get(name)
	if !ok {
		return nil, nil
	}

	if ret, ok := val.([]int); ok {
		return ret, nil
	}

	return nil, fmt.Errorf("could not convert %T{%v} to []int for %s", val, val, name)
}

func (c *Config) Generic(name string) (cli.Generic, error) {
	val, ok := c.Get(name)
	if !ok {
		return nil, nil
	}

	if ret, ok := val.(cli.Generic); ok {
		return ret, nil
	}

	return JSON{&val}, nil
}

func (c *Config) Bool(name string) (bool, error) {
	val, ok := c.Get(name)
	if !ok {
		return false, nil
	}

	if ret, ok := val.(bool); ok {
		return ret, nil
	}

	return false, fmt.Errorf("could not convert %T{%v} to bool for %s", val, val, name)
}

func (c *Config) BoolT(name string) (bool, error) {
	val, ok := c.Get(name)
	if !ok {
		return true, nil
	}

	if ret, ok := val.(bool); ok {
		return ret, nil
	}

	return true, fmt.Errorf("could not convert %T{%v} to boolT for %s", val, val, name)
}
