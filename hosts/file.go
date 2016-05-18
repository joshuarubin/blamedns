package hosts

import "regexp"

var fileRe = regexp.MustCompile(`\S+\s+(\S+)`)

var FileParser = ParserFunc(func(text string) (host string, valid bool) {
	res := fileRe.FindStringSubmatch(text)
	if len(res) < 2 {
		return
	}

	return res[1], true
})
