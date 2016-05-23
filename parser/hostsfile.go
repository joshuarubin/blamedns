package parser

import (
	"jrubin.io/blamedns/textmodifier"

	"github.com/Sirupsen/logrus"
)

type HostsFileParser struct {
	HostAdder HostAdder
	Logger    *logrus.Logger
}

func (h HostsFileParser) Reset(fileName string) {
	h.HostAdder.Reset(fileName)
}

func (h HostsFileParser) Parse(fileName string, lineNum int, text string) (ret bool) {
	textmodifier.New(&text).StripComments().ExtractField(1).ToLower().UnFQDN().StripPort()

	if ret = ValidateHost(h.Logger, fileName, lineNum, text); ret {
		h.HostAdder.AddHost(fileName, text)
	}

	return
}
