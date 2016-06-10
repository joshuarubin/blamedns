package config

import "strings"

// StringSlice is an opaque type for []string to satisfy flag.Value
type StringSlice []string

// Value returns the slice of strings set by this flag
func (f *StringSlice) Value() []string {
	return *f
}

// String returns a readable representation of this value (for usage defaults)
func (f *StringSlice) String() string {
	return strings.Join(*f, ",")
}

// Set the string value to the comma separated list of values
func (f *StringSlice) Set(value string) error {
	*f = []string{}
	for _, s := range strings.Split(value, ",") {
		s = strings.TrimSpace(s)
		*f = append(*f, s)
	}
	return nil
}
