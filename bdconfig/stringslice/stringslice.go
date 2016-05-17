package stringslice

import (
	"fmt"
	"io"
)

func Write(name string, ss []string, w io.Writer) (int, error) {
	if len(ss) == 0 {
		return 0, nil
	}

	n, err := fmt.Fprintf(w, "%s = [\n", name)
	if err != nil {
		return n, err
	}

	for _, h := range ss {
		var o int
		o, err = fmt.Fprintf(w, "  \"%s\",\n", h)
		n += o
		if err != nil {
			return n, err
		}
	}

	o, err := fmt.Fprintf(w, "]\n")
	n += o

	return n, err
}
