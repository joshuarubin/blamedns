package dnscachetypes

import (
	"time"

	"github.com/miekg/dns"
)

type RR struct {
	dns.RR
	Expires time.Time
}

func (m RR) Equal(d *RR) bool {
	if m.Header().Name != d.Header().Name {
		return false
	}

	if m.Header().Rrtype != d.Header().Rrtype {
		return false
	}

	// no need to test class since we only use ClassINET
	// don't compare TTL or Expires either

	// this only works if the TTL is set to 0 on both
	return m.String() == d.String()
}

func (m RR) TTL() TTL {
	return TTL(m.Expires.Sub(time.Now().UTC()))
}

func (m *RR) UpdateTTL() {
	m.Header().Ttl = m.TTL().Seconds()
}

func (m RR) Expired() bool {
	return m.Expires.Before(time.Now().UTC())
}

func NewRR(r dns.RR) *RR {
	ttl := time.Duration(r.Header().Ttl) * time.Second
	r = dns.Copy(r)
	r.Header().Ttl = 0 // set to 0 so that Equal works using r.String()
	return &RR{
		RR:      r,
		Expires: time.Now().UTC().Add(ttl),
	}
}
