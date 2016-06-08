package config

import "encoding/json"

type JSON struct {
	Data interface{}
}

func (f JSON) Set(value string) error {
	return f.UnmarshalJSON([]byte(value))
}

func (f JSON) String() string {
	b, _ := f.MarshalJSON()
	return string(b)
}

func (f JSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Data)
}

func (f JSON) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &f.Data)
}
