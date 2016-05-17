package cliconfig

import (
	"fmt"
	"reflect"

	"github.com/codegangsta/cli"
)

// Parse the cli.Context and set the values in dest appropriately
func (cc CLIConfig) Parse(c *cli.Context, dest interface{}) error {
	return structForEach("", reflect.ValueOf(dest), func(fieldName string, _, fieldValue reflect.Value, envVar, usage string, hidden bool) error {
		fai := fieldValue.Addr().Interface()

		if um, ok := fai.(Unmarshaler); ok {
			val, err := um.UnmarshalCLIConfig(c.String(fieldName))
			if err != nil {
				return err
			}

			cc.set(fai, val, fieldName)
			return nil
		}

		switch fieldValue.Kind() {
		case reflect.Bool:
			fieldValue.SetBool(c.Bool(fieldName))
		case reflect.Int:
			cc.set(fai, c.Int(fieldName), fieldName)
		case reflect.Float64:
			cc.set(fai, c.Float64(fieldName), fieldName)
		case reflect.String:
			cc.set(fai, c.String(fieldName), fieldName)
		case reflect.Int64: // really only looking for time.Duration
			if fieldValue.Type() != timeDurationType {
				panic(fmt.Sprintf("CLIConfig.Parse: invalid field type: %s (%s)", fieldValue.Kind().String(), fieldName))
			}
			cc.set(fai, c.Duration(fieldName), fieldName)
		case reflect.Slice: // looking for []int and []string
			switch fieldValue.Type() {
			case intSliceType:
				var dflt []int
				if iface, ok := cc.defaults[fieldName]; ok {
					dflt = iface.([]int)
				}
				n := len(dflt)

				vals := c.IntSlice(fieldName)

				// strip out the default values if the user supplied any
				if len(vals) > n {
					vals = vals[n:]
				}

				cc.set(fai, vals, fieldName)
			case stringSliceType:
				var dflt []string
				if iface, ok := cc.defaults[fieldName]; ok {
					dflt = iface.([]string)
				}
				n := len(dflt)

				allVals := c.StringSlice(fieldName)

				// strip out the default values if the user supplied any
				if len(allVals) > n {
					allVals = allVals[n:]
				}

				// strip empty values
				var vals []string
				for _, v := range allVals {
					if len(v) > 0 {
						vals = append(vals, v)
					}
				}

				cc.set(fai, vals, fieldName)
			default:
				panic(fmt.Sprintf("CLIConfig.Parse: invalid field type: %s (%s)", fieldValue.Kind().String(), fieldName))
			}
		default:
			panic(fmt.Sprintf("CLIConfig.Parse: invalid field type: %s (%s)", fieldValue.Kind().String(), fieldName))
		}

		return nil
	})
}
