package parser

import (
	"github.com/Sirupsen/logrus"
	"jrubin.io/blamedns/textmodifier"
)

var _ Parser = DomainParser{}

type DomainParser struct {
	HostAdder HostAdder
	Logger    *logrus.Logger
}

func (d DomainParser) Parse(fileName string, lineNum int, text string) (ret bool) {
	textmodifier.New(&text).StripComments().TrimSpace().ToLower().UnFQDN().StripPort()

	if ret = ValidateHost(d.Logger, fileName, lineNum, text); ret {
		d.HostAdder.AddHost(fileName, text)
	}

	return
}
