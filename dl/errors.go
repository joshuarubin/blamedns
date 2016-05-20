package dl

import (
	"fmt"
	"net/http"
)

type ErrStatusCode int

func (e ErrStatusCode) Error() string {
	return fmt.Sprintf("unexpected status code: %d (%s)", e.Int(), e.String())
}

func (e ErrStatusCode) String() string {
	return http.StatusText(e.Int())
}

func (e ErrStatusCode) Int() int {
	return int(e)
}

func NewErrStatusCode(i int) error {
	return ErrStatusCode(i)
}
