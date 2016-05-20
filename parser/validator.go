package parser

import (
	"strings"

	"github.com/Sirupsen/logrus"
)

func ValidateHost(logger *logrus.Logger, fileName string, ln int, text string) bool {
	if len(text) == 0 {
		return false
	}

	// the entire hostname (including the delimiting dots but not a trailing
	// dot) has a maximum of 253 ascii characters
	if len(text) > 253 {
		logger.Warnf("%s:%d: hostname too long: %s", fileName, ln, text)
		return false
	}

	labels := strings.Split(text, ".")
	for _, label := range labels {
		// each label must be between 1 and 63 characters long
		l := len(label)

		if l == 0 {
			logger.Warnf("%s:%d: empty label: %s", fileName, ln, text)
			return false
		}

		if l > 63 {
			logger.Warnf("%s:%d: label too long: %s (%s)", fileName, ln, label, text)
			return false
		}

		// hostname labels may contain only the ASCII letters 'a' through 'z' (in a
		// case-insensitive manner), the digits '0' through '9', and the hyphen
		// ('-')
		for i := 0; i < l; i++ {
			c := label[i]

			if c >= 'a' && c <= 'z' {
				continue
			}

			// if c >= 'A' && c <= 'Z' {
			// 	continue
			// }

			if c >= '0' && c <= '9' {
				continue
			}

			if c == '_' {
				continue
			}

			// must not start or end with a hyphen
			if c == '-' && i > 0 && i < l-1 {
				continue
			}

			logger.Warnf("%s:%d: invalid character \"%c\" (%s)", fileName, ln, c, text)
			return false
		}
	}

	return true
}
