package textmodifier

import (
	"regexp"
	"strings"

	"github.com/miekg/dns"
)

type Modifier struct {
	Text *string
}

func New(text *string) *Modifier {
	return &Modifier{Text: text}
}

func (m *Modifier) String() string {
	return *m.Text
}

func (m *Modifier) TrimSpace() *Modifier {
	*m.Text = strings.TrimSpace(*m.Text)
	return m
}

func (m *Modifier) ToLower() *Modifier {
	*m.Text = strings.ToLower(*m.Text)
	return m
}

func (m *Modifier) StripComments() *Modifier {
	i := strings.Index(*m.Text, "#")
	if i == 0 {
		*m.Text = ""
		return m
	}

	if i != -1 {
		*m.Text = (*m.Text)[:i-1]
	}

	return m
}

var portRe = regexp.MustCompile(`(\S+):\d+`)

func (m *Modifier) StripPort() *Modifier {
	if res := portRe.FindStringSubmatch(*m.Text); len(res) > 1 {
		*m.Text = res[1]
	}
	return m
}

func (m *Modifier) UnFQDN() *Modifier {
	if dns.IsFqdn(*m.Text) {
		*m.Text = (*m.Text)[:len(*m.Text)-1]
	}
	return m
}

func (m *Modifier) ExtractField(num int) *Modifier {
	if f := strings.Fields(*m.Text); num < len(f) {
		*m.Text = f[num]
	} else {
		*m.Text = ""
	}
	return m
}
