package cliconfig

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// largely copied from github.com/fatih/camelcase
func fromCamel(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			if lastClass == 0 {
				class = 3
			} else {
				class = lastClass
			}
		default:
			class = 4
		}

		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}

		lastClass = class
	}

	for i := 0; i < len(runes)-1; i++ {
		// handle upper case -> lower case sequences, e.g.
		// "PDFL", "oader" -> "PDF", "Loader"
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	for i := 0; i < len(runes)-1; i++ {
		if len(runes[i]) == 0 {
			continue
		}

		var j int
		for j = i + 1; j < len(runes); j++ {
			if len(runes[j]) > 0 {
				break
			}
		}

		// handle leading numbers e.g.
		// "99" "Bottles" -> "99Bottles"
		if unicode.IsDigit(runes[i][0]) {
			runes[i] = append(runes[i], runes[j]...)
			runes[j] = []rune{}
		}
	}

	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}

func toCamel(words []string) string {
	var ret string
	for i, word := range words {
		if i == 0 {
			// don't change case of first word as it indicates public/private in Go
			ret += word
			continue
		}

		r, size := utf8.DecodeRuneInString(word)
		if r == utf8.RuneError {
			continue
		}

		ret += string(unicode.ToUpper(r))
		ret += word[size:]
	}
	return ret
}

func toSpinal(words []string) string {
	var ret []string
	for _, word := range words {
		ret = append(ret, strings.ToLower(word))
	}
	return strings.Join(ret, "-")
}

func toUpperSnake(words []string) string {
	var ret []string
	for _, word := range words {
		ret = append(ret, strings.ToUpper(word))
	}
	return strings.Join(ret, "_")
}

func massageName(fn func([]string) string, partsParts ...[]string) string {
	var data []string
	for _, parts := range partsParts {
		data = append(data, parts...)
	}
	return fn(data)
}
