package parser

type HostAdder interface {
	AddHost(source, host string)
	Reset(source string)
}

type Parser interface {
	Parse(fileName string, lineNum int, line string) bool
	Reset(fileName string)
}
