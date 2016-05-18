package hosts

var DomainParser = ParserFunc(func(text string) (host string, valid bool) {
	return text, true
})
