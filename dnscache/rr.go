package dnscachetypes

import (
	"time"

	"github.com/miekg/dns"
)

type RR struct {
	rr      dns.RR
	Expires time.Time
}

func (m RR) Header() *dns.RR_Header {
	return m.rr.Header()
}

func (m RR) String() string {
	return m.rr.String()
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

func (m RR) RR() dns.RR {
	r := dns.Copy(m.rr)
	r.Header().Ttl = m.TTL().Seconds()
	return r
}

func (m RR) Expired() bool {
	return m.Expires.Before(time.Now().UTC())
}

func NewRR(r dns.RR) *RR {
	ttl := time.Duration(r.Header().Ttl) * time.Second
	r = dns.Copy(r)
	r.Header().Ttl = 0 // set to 0 so that Equal works using r.String()
	return &RR{
		rr:      r,
		Expires: time.Now().UTC().Add(ttl),
	}
}
