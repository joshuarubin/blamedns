package hosts

type Parser interface {
	Parse(string) (string, bool)
}

type ParserFunc func(string) (string, bool)

func (f ParserFunc) Parse(text string) (string, bool) {
	return f(text)
}
