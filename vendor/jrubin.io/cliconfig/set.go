package cliconfig

import "reflect"

func isInitial(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func (cc CLIConfig) set(dest, value interface{}, fieldName string) {
	v := reflect.ValueOf(value)

	// if value is go's "initial value, don't write it
	if isInitial(v) {
		return
	}

	cur := reflect.Indirect(reflect.ValueOf(dest))

	dflt, hasDefault := cc.defaults[fieldName]

	var isDefault bool
	if hasDefault {
		if eq, ok := value.(Equaler); ok {
			isDefault = eq.Equal(dflt)
		} else {
			isDefault = reflect.DeepEqual(value, dflt)
		}
	}

	// if the value in cur is go's "initial" value or isDefault is false it is
	// safe to overwrite
	if !isDefault || isInitial(cur) {
		cur.Set(v)
	}
}
