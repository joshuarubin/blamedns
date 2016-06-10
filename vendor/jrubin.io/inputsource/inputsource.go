package inputsource

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var _ altsrc.InputSourceContext = &InputSource{}

type InputSource struct {
	data reflect.Value
}

func New(data interface{}) *InputSource {
	value := reflect.Indirect(reflect.ValueOf(data))

	if value.Kind() != reflect.Struct {
		panic("inputsource: structure is not a struct")
	}

	return &InputSource{
		data: value,
	}
}

func getField(v reflect.Value, key string) (interface{}, bool) {
	// iterate through key name like "a-b-c" as:
	// 1. "a-b-c"
	// 2. "a-b"
	// 3. "a"
RECURSE:
	for {
		for i := len(key); i != -1; i = strings.LastIndex(key[:i], "-") {
			// strip out "-"
			test := strings.ToLower(strings.Replace(key[:i], "-", "", -1))

			// find field case-insensitively
			f := v.FieldByNameFunc(func(name string) bool {
				return test == strings.ToLower(name)
			})

			if !f.IsValid() {
				// field not found
				continue
			}

			f = reflect.Indirect(f)

			if i == len(key) {
				// found complete match, return regardless of type
				return f.Interface(), true
			}

			if f.Type().Kind() == reflect.Struct {
				// found struct match for partial key name, recurse
				v = f
				key = key[len(test)+1:]
				continue RECURSE
			}
		}

		panic(fmt.Errorf("could not find key: %s", key))
	}
}

func (s *InputSource) get(key string) (interface{}, bool) {
	return getField(s.data, key)
}

func (s *InputSource) Int(name string) (int, error) {
	val, ok := s.get(name)
	if !ok {
		return 0, nil
	}

	if ret, ok := val.(int); ok {
		return ret, nil
	}

	return 0, fmt.Errorf("could not convert %T{%v} to int for %s", val, val, name)
}

func (s *InputSource) Duration(name string) (time.Duration, error) {
	val, ok := s.get(name)
	if !ok {
		return 0, nil
	}

	if ret, ok := val.(time.Duration); ok {
		return ret, nil
	}

	return 0, fmt.Errorf("could not convert %T{%v} to time.Duration for %s", val, val, name)
}

func (s *InputSource) Float64(name string) (float64, error) {
	val, ok := s.get(name)
	if !ok {
		return 0, nil
	}

	if ret, ok := val.(float64); ok {
		return ret, nil
	}

	return 0, fmt.Errorf("could not convert %T{%v} to float64 for %s", val, val, name)
}

func (s *InputSource) String(name string) (string, error) {
	val, ok := s.get(name)
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

func (s *InputSource) StringSlice(name string) ([]string, error) {
	val, ok := s.get(name)
	if !ok {
		return nil, nil
	}

	if ret, ok := val.([]string); ok {
		return ret, nil
	}

	return nil, fmt.Errorf("could not convert %T{%v} to []string for %s", val, val, name)
}

func (s *InputSource) IntSlice(name string) ([]int, error) {
	val, ok := s.get(name)
	if !ok {
		return nil, nil
	}

	if ret, ok := val.([]int); ok {
		return ret, nil
	}

	return nil, fmt.Errorf("could not convert %T{%v} to []int for %s", val, val, name)
}

type Genericer interface {
	Generic() cli.Generic
}

func (s *InputSource) Generic(name string) (cli.Generic, error) {
	val, ok := s.get(name)
	if !ok {
		return nil, nil
	}

	if ret, ok := val.(cli.Generic); ok {
		return ret, nil
	}

	if ret, ok := val.(Genericer); ok {
		return ret.Generic(), nil
	}

	return nil, fmt.Errorf("could not convert %T{%v} to cli.Generic for %s", val, val, name)
}

func (s *InputSource) Bool(name string) (bool, error) {
	val, ok := s.get(name)
	if !ok {
		return false, nil
	}

	if ret, ok := val.(bool); ok {
		return ret, nil
	}

	return false, fmt.Errorf("could not convert %T{%v} to bool for %s", val, val, name)
}

func (s *InputSource) BoolT(name string) (bool, error) {
	val, ok := s.get(name)
	if !ok {
		return true, nil
	}

	if ret, ok := val.(bool); ok {
		return ret, nil
	}

	return true, fmt.Errorf("could not convert %T{%v} to boolT for %s", val, val, name)
}
