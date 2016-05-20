package parser

type HostAdder interface {
	AddHost(source, host string)
}
type Parser interface {
	Parse(fileName string, lineNum int, line string) bool
}

type ParserFunc func(fileName string, lineNum int, line string) bool

func (f ParserFunc) Parse(fileName string, lineNum int, line string) bool {
	return f(fileName, lineNum, line)
}

var _ Parser = ParserFunc(func(string, int, string) bool { return false })
