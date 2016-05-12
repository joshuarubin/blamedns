package main

import "github.com/codegangsta/cli"

type stringSlice struct {
	flag
	Default cli.StringSlice
	Values  cli.StringSlice
}

func newStringSlice(s ...string) stringSlice {
	ret := stringSlice{
		Default: s,
	}

	ret.Values = make(cli.StringSlice, len(ret.Default))
	copy(ret.Values, ret.Default)

	return ret
}

func (ss *stringSlice) parseFlag(c *cli.Context) error {
	ss.Values = c.StringSlice(ss.Name)

	// strip out the default values if the user supplied any
	if len(ss.Values) > len(ss.Default) {
		ss.Values = ss.Values[len(ss.Default):]
	}

	ss.Values = stripEmpty(ss.Values)

	return nil
}

type flag struct {
	Name   string
	Usage  string
	EnvVar string
	Hidden bool
}

func (ss *stringSlice) Flag(c flag) cli.Flag {
	ss.flag = c
	return cli.StringSliceFlag{
		Name:   c.Name,
		EnvVar: c.EnvVar,
		Value:  &ss.Values,
		Usage:  c.Usage,
	}
}

func (ss stringSlice) IsDefault() bool {
	if ss.Values == nil && ss.Default == nil {
		return true
	}

	if ss.Values == nil || ss.Default == nil {
		return false
	}

	if len(ss.Values) != len(ss.Default) {
		return false
	}

	for i := range ss.Values {
		if ss.Values[i] != ss.Default[i] {
			return false
		}
	}

	return true
}

func stripEmpty(s []string) []string {
	var ret []string

	for _, v := range s {
		if len(v) > 0 {
			ret = append(ret, v)
		}
	}

	return ret
}
