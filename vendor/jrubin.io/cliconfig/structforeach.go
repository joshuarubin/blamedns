package cliconfig

import (
	"reflect"
	"strings"
)

const tagName = "cli"

func structForEach(prefix string, structure reflect.Value, fn structFn) error {
	value := reflect.Indirect(structure)

	if value.Kind() != reflect.Struct {
		panic("cliconfig: structure is not a struct")
	}

	prefixParts := fromCamel(prefix)

	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get(tagName)

		// omit fields with "-" tag
		if tag == "-" {
			continue
		}

		name := tag
		hidden := false
		var usage string

		ci := strings.Index(tag, ",")
		if ci != -1 {
			name = tag[:ci]
			tval := tag[ci+1:]
			if tval == "hidden" {
				hidden = true
			} else {
				usage = tval
			}
		}

		if len(name) == 0 {
			name = typ.Field(i).Name
		}
		fieldValue := value.Field(i)

		nameParts := fromCamel(name)
		fieldName := massageName(toSpinal, prefixParts, nameParts)
		envVar := massageName(toUpperSnake, prefixParts, nameParts)

		dflt := fieldValue

		dfn, hasDefault := fieldValue.Interface().(CustomType)
		if hasDefault {
			dflt = reflect.ValueOf(dfn.Default(typ.Field(i).Name))
		}

		if !hasDefault && fieldValue.Kind() == reflect.Struct {
			if err := structForEach(massageName(toCamel, prefixParts, nameParts), fieldValue, fn); err != nil {
				return err
			}
		} else {
			if err := fn(fieldName, dflt, fieldValue, envVar, usage, hidden); err != nil {
				return err
			}
		}
	}

	return nil
}
