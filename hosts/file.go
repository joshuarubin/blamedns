package hosts

import (
	"regexp"
	"strings"
)

var fileRe = regexp.MustCompile(`\S+\s+(\S+)`)

var FileParser = ParserFunc(func(text string) (host string, valid bool) {
	i := strings.Index(text, "#")
	if i == 0 {
		return
	}

	if i != -1 {
		text = text[:i-1]
	}

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return
	}

	res := fileRe.FindStringSubmatch(text)
	if len(res) < 2 {
		return
	}

	return res[1], true
})
