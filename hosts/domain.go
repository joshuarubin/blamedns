package hosts

import "strings"

var DomainParser = ParserFunc(func(text string) (host string, valid bool) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return
	}

	return text, true
})
