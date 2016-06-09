package config

import (
	"flag"
	"os"
	"strings"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var _ cli.Flag = &StringSliceFlag{}
var _ altsrc.FlagInputSourceExtension = &StringSliceFlag{}

// StringSliceFlag is a string flag that can be specified multiple times on the
// command-line
type StringSliceFlag struct {
	Name        string
	Value       cli.StringSlice
	Destination *cli.StringSlice
	Usage       string
	EnvVar      string
	Hidden      bool
	set         *flag.FlagSet
}

// String returns the usage
func (f StringSliceFlag) String() string {
	return cli.FlagStringer(cli.StringSliceFlag{
		Name:   f.Name,
		Value:  &f.Value,
		Usage:  f.Usage,
		EnvVar: f.EnvVar,
		Hidden: f.Hidden,
	})
}

// Apply populates the flag given the flag set and environment
func (f *StringSliceFlag) Apply(set *flag.FlagSet) {
	f.set = set

	if f.Destination == nil {
		f.Destination = &cli.StringSlice{}
	} else {
		*f.Destination = cli.StringSlice{}
	}

	if f.EnvVar != "" {
		for _, envVar := range strings.Split(f.EnvVar, ",") {
			envVar = strings.TrimSpace(envVar)
			if envVal := os.Getenv(envVar); envVal != "" {
				newVal := cli.StringSlice{}
				for _, s := range strings.Split(envVal, ",") {
					s = strings.TrimSpace(s)
					newVal.Set(s)
				}
				*f.Destination = newVal
				break
			}
		}
	}

	eachName(f.Name, func(name string) {
		set.Var(f.Destination, name, f.Usage)
	})
}

// GetName returns the name of a flag.
func (f StringSliceFlag) GetName() string {
	return f.Name
}

func eachName(longName string, fn func(string)) {
	parts := strings.Split(longName, ",")
	for _, name := range parts {
		name = strings.Trim(name, " ")
		fn(name)
	}
}

// ApplyInputSourceValue applies a StringSlice value to the flagSet if required
func (f *StringSliceFlag) ApplyInputSourceValue(context *cli.Context, isc altsrc.InputSourceContext) error {
	if f.set != nil {
		if !context.IsSet(f.Name) && !isEnvVarSet(f.EnvVar) {
			value, err := isc.StringSlice(f.Name)
			if err != nil {
				return err
			}
			if value != nil {
				var sliceValue cli.StringSlice = value
				eachName(f.Name, func(name string) {
					underlyingFlag := f.set.Lookup(f.Name)
					if underlyingFlag != nil {
						if ss, ok := underlyingFlag.Value.(*cli.StringSlice); ok {
							*ss = sliceValue
						} else {
							underlyingFlag.Value = &sliceValue
						}
					}
				})
			}
		}
	}
	return nil
}

func isEnvVarSet(envVars string) bool {
	for _, envVar := range strings.Split(envVars, ",") {
		envVar = strings.TrimSpace(envVar)
		if envVal := os.Getenv(envVar); envVal != "" {
			if len(envVal) > 0 {
				return true
			}
		}
	}

	return false
}
