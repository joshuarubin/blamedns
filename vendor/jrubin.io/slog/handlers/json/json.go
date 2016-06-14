// Package json implements a JSON handler.
package json

import (
	j "encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"jrubin.io/slog"
)

// Default handler outputting to stderr.
var Default = New(os.Stderr)

// Logger returns a logger configured to output json at level or higher to
// stderr.
func Logger(level slog.Level) *slog.Logger {
	return slog.New().RegisterHandler(level, Default)
}

// Handler implementation.
type Handler struct {
	mu  sync.Mutex
	enc *j.Encoder
}

// New handler.
func New(w io.Writer) *Handler {
	return &Handler{
		enc: j.NewEncoder(w),
	}
}

type entry struct {
	Fields  slog.Fields `json:"fields"`
	Level   slog.Level  `json:"level"`
	Time    time.Time   `json:"time"`
	Message string      `json:"msg"`
}

func newEntry(e *slog.Entry) *entry {
	ret := &entry{
		Fields:  slog.Fields{},
		Level:   e.Level,
		Time:    e.Time,
		Message: e.Message,
	}

	for key, value := range e.Fields {
		switch value := value.(type) {
		case fmt.Stringer:
			ret.Fields[key] = value.String()
		case error:
			ret.Fields[key] = value.Error()
		default:
			ret.Fields[key] = value
		}
	}

	return ret
}

// HandleLog implements slog.Handler.
func (h *Handler) HandleLog(e *slog.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.enc.Encode(newEntry(e))
}
