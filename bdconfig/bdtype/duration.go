package bdtype

import (
	"encoding"
	"fmt"
	"time"

	"jrubin.io/cliconfig"
)

const (
	defaultBlockTTL           = 1 * time.Hour
	defaultUpdateInterval     = 24 * time.Hour
	defaultDNSTimeout         = 5 * time.Second
	defaultDNSInterval        = 200 * time.Millisecond
	defaultCachePruneInterval = 1 * time.Hour
)

type Duration time.Duration

var (
	tmpDur                          = Duration(0)
	_      cliconfig.CustomType     = tmpDur
	_      encoding.TextMarshaler   = tmpDur
	_      encoding.TextUnmarshaler = &tmpDur
)

func (d Duration) Default(name string) interface{} {
	switch name {
	case "dns-block-ttl":
		return defaultBlockTTL
	case "dns-block-update-interval":
		return defaultUpdateInterval
	case "dns-timeout":
		return defaultDNSTimeout
	case "dns-interval":
		return defaultDNSInterval
	case "dns-cache-prune-interval":
		return defaultCachePruneInterval
	}
	panic(fmt.Sprintf("bdtype.Duration.Default unknown name: %s", name))
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return d.Duration().String()
}

func (d Duration) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret Duration
	err := ret.UnmarshalText([]byte(text))
	return ret, err
}

func (d Duration) Equal(val interface{}) bool {
	if tval, ok := val.(time.Duration); ok {
		return d.Duration() == tval
	}
	return false
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(tmp)
	return nil
}
