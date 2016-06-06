// Package auto implements a handler that will use json for files and text for
// terminals.
package auto

import (
	"io"
	"os"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/json"
	"jrubin.io/slog/handlers/text"
)

// Default handler outputting to stderr.
var Default = New(os.Stderr)

// Handler implementation.
type Handler struct {
	handler slog.Handler
}

// New handler.
func New(w io.Writer) *Handler {
	if w == os.Stdout || w == os.Stderr {
		return &Handler{
			handler: text.New(w),
		}
	}

	return &Handler{
		handler: json.New(w),
	}
}

// HandleLog implements slog.Handler.
func (h *Handler) HandleLog(e *slog.Entry) error {
	return h.handler.HandleLog(e)
}
