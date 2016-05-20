package parser

import (
	"regexp"

	"jrubin.io/blamedns/textmodifier"

	"github.com/Sirupsen/logrus"
)

var (
	_      Parser = HostsFileParser{}
	fileRe        = regexp.MustCompile(`\S+\s+(\S+)`)
)

type HostsFileParser struct {
	HostAdder HostAdder
	Logger    *logrus.Logger
}

func (h HostsFileParser) Parse(fileName string, lineNum int, text string) (ret bool) {
	textmodifier.New(&text).StripComments().ExtractField(1).ToLower().UnFQDN().StripPort()

	if ret = ValidateHost(h.Logger, fileName, lineNum, text); ret {
		h.HostAdder.AddHost(fileName, text)
	}

	return
}
