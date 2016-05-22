package blocker

import "sort"

type sources []string

func (s *sources) Add(source string) {
	n := sort.SearchStrings(*s, source)

	if n < len(*s) && (*s)[n] == source {
		// already exists
		return
	}

	// insert source into slice at index n
	*s = append(*s, "")
	copy((*s)[n+1:], (*s)[n:])
	(*s)[n] = source
}

func (s *sources) Remove(source string) bool {
	if s == nil {
		return false
	}

	n := sort.SearchStrings(*s, source)

	if n < len(*s) && (*s)[n] == source {
		*s = append((*s)[:n], (*s)[n+1:]...)
		return true
	}

	return false
}

func (s *sources) Has(source string) bool {
	if s == nil {
		return false
	}

	n := sort.SearchStrings(*s, source)

	if n < len(*s) && (*s)[n] == source {
		return true
	}

	return false
}

func (s *sources) Len() int {
	return len(*s)
}

func newSources(s ...string) *sources {
	ret := sources(s)
	return &ret
}
