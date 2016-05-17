package cliconfig

import (
	"reflect"
	"time"
)

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
	stringSliceType  = reflect.TypeOf([]string{})
	intSliceType     = reflect.TypeOf([]int{})
)

type CustomType interface {
	Default(string) interface{}
	UnmarshalCLIConfig(string) (interface{}, error)
	Equal(interface{}) bool
}

type structFn func(fieldName string, dflt, fieldValue reflect.Value, envVar, usage string, hidden bool) error

// A CLIConfig is used to turn a struct into cli.Flags and then parse a
// cli.Context back into the struct. Prefix is prepended to all flag names and
// environment variables. If Structure has any values, they will be used as the
// default. Structure may also define the "cli" struct tag as follows:
//
// `cli:"field-name" // uses field-name instead of the struct field's name for
// the arguments and environment variables
//
// `cli:"field-name,hidden" // will use field-name and set the cli.Flag.Hidden
// field to true
//
// `cli:",hidden" // will not change the field name, but will set the
// cli.Flag.Hidden field to true
//
// `cli:"field-name,usage text goes here" // will change the field name and
// supply usage text
//
// `cli:",usage text goes here" // will not change the field name, but will
// supply usage text
type CLIConfig struct {
	Structure interface{}
	defaults  map[string]interface{}
}

// New allocates a CLIConfig
func New(structure interface{}) *CLIConfig {
	return &CLIConfig{
		Structure: structure,
	}
}
