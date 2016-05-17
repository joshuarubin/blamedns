package cliconfig

import (
	"fmt"
	"reflect"
	"time"

	"github.com/codegangsta/cli"
)

// Flags returns a slice of cli.Flag suitable for using with a cli.App or
// cli.Command
func (cc *CLIConfig) Flags() []cli.Flag {
	var ret []cli.Flag

	if cc.defaults == nil {
		cc.defaults = map[string]interface{}{}
	}

	_ = structForEach("", reflect.ValueOf(cc.Structure), func(fieldName string, dflt, _ reflect.Value, envVar, usage string, hidden bool) error {
		switch dflt.Kind() {
		case reflect.Bool:
			ret = append(ret, cli.BoolFlag{
				Name:   fieldName,
				EnvVar: envVar,
				Usage:  usage,
				Hidden: hidden,
			})
		case reflect.Int:
			val := int(dflt.Int())
			cc.defaults[fieldName] = val
			ret = append(ret, cli.IntFlag{
				Name:   fieldName,
				EnvVar: envVar,
				Value:  val,
				Usage:  usage,
				Hidden: hidden,
			})
		case reflect.Float64:
			val := dflt.Float()
			cc.defaults[fieldName] = val
			ret = append(ret, cli.Float64Flag{
				Name:   fieldName,
				EnvVar: envVar,
				Value:  val,
				Usage:  usage,
				Hidden: hidden,
			})
		case reflect.String:
			val := dflt.String()
			cc.defaults[fieldName] = val
			ret = append(ret, cli.StringFlag{
				Name:   fieldName,
				EnvVar: envVar,
				Value:  val,
				Usage:  usage,
				Hidden: hidden,
			})
		case reflect.Int64: // really only looking for time.Duration
			if dflt.Type() != timeDurationType {
				panic(fmt.Sprintf("CLIConfig.Flags: invalid field type: %s (%s)", dflt.Kind().String(), fieldName))
			}
			val := time.Duration(dflt.Int())
			cc.defaults[fieldName] = val
			ret = append(ret, cli.DurationFlag{
				Name:   fieldName,
				EnvVar: envVar,
				Value:  val,
				Usage:  usage,
				Hidden: hidden,
			})
		case reflect.Slice: // looking for []int and []string
			switch dflt.Type() {
			case intSliceType:
				is := dflt.Interface().([]int)
				cc.defaults[fieldName] = is
				ci := cli.IntSlice(is)

				ret = append(ret, cli.IntSliceFlag{
					Name:   fieldName,
					EnvVar: envVar,
					Value:  &ci,
					Usage:  usage,
					Hidden: hidden,
				})
			case stringSliceType:
				ss := dflt.Interface().([]string)
				cc.defaults[fieldName] = ss
				cs := cli.StringSlice(ss)

				ret = append(ret, cli.StringSliceFlag{
					Name:   fieldName,
					EnvVar: envVar,
					Value:  &cs,
					Usage:  usage,
					Hidden: hidden,
				})
			default:
				panic(fmt.Sprintf("CLIConfig.Flags: invalid field type: %s (%s)", dflt.Kind().String(), fieldName))
			}
		default:
			panic(fmt.Sprintf("CLIConfig.Flags: invalid field type: %s (%s)", dflt.Kind().String(), fieldName))
		}

		return nil
	})

	return ret
}
