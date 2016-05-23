package dnscache

import "time"

type TTL time.Duration

func (t TTL) Duration() time.Duration {
	return time.Duration(t)
}

func (t TTL) Seconds() uint32 {
	return uint32(t.Duration() / time.Second)
}
