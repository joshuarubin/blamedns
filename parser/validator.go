package parser

import (
	"strings"

	"jrubin.io/slog"
)

func ValidateHost(logger slog.Interface, fileName string, ln int, text string) bool {
	if len(text) == 0 {
		return false
	}

	ctxLog := logger.WithFields(slog.Fields{
		"file": fileName,
		"line": ln,
		"host": text,
	})

	// the entire hostname (including the delimiting dots but not a trailing
	// dot) has a maximum of 253 ascii characters
	if len(text) > 253 {
		ctxLog.Warn("hostname too long")
		return false
	}

	labels := strings.Split(text, ".")
	for _, label := range labels {
		// each label must be between 1 and 63 characters long
		l := len(label)

		if l == 0 {
			ctxLog.Warn("hostname has empty label")
			return false
		}

		if l > 63 {
			ctxLog.WithField("label", label).Warn("hostname has label that's too long")
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

			ctxLog.WithField("char", c).Warn("hostname has invalid character")
			return false
		}
	}

	return true
}
