package parser

import (
	"jrubin.io/blamedns/textmodifier"
	"jrubin.io/slog"
)

type DomainParser struct {
	HostAdder HostAdder
	Logger    slog.Interface
}

func (d DomainParser) Reset(fileName string) {
	d.HostAdder.Reset(fileName)
}

func (d DomainParser) Parse(fileName string, lineNum int, text string) (ret bool) {
	textmodifier.New(&text).StripComments().TrimSpace().ToLower().UnFQDN().StripPort()

	if ret = ValidateHost(d.Logger, fileName, lineNum, text); ret {
		d.HostAdder.AddHost(fileName, text)
	}

	return
}
