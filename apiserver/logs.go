package apiserver

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"jrubin.io/slog"

	"golang.org/x/net/websocket"
)

// A Leveler is used to determine from an http.Request the logLevel that should
// be used for the websocket.
type Leveler interface {
	Level(req *http.Request) slog.Level
}

// LevelerFunc is a convenience wrapper to create a Leveler.
type LevelerFunc func(*http.Request) slog.Level

// Level implements the Leveler interface.
func (f LevelerFunc) Level(req *http.Request) slog.Level {
	return f(req)
}

// LogsHandler returns an http.Handler that will push JSON formatted log
// messages to the client. Handler should only be called once per logger. By
// default, each request is configured to receive logs at defaultLevel or
// higher. If you would like to be able to set a custom logLevel per websocket
// connection, use the Leveler interface to determine what level the request
// should use.
func LogsHandler(logger *slog.Logger, defaultLevel slog.Level, leveler Leveler) http.Handler {
	wsLogger := newWSLogHook(logger)
	logger.RegisterHandler(slog.DebugLevel, wsLogger)

	return websocket.Handler(func(ws *websocket.Conn) {
		level := defaultLevel
		if leveler != nil {
			level = leveler.Level(ws.Request())
		}

		<-wsLogger.addConn(ws, level)
	})
}

type chanLevel struct {
	ch    chan struct{}
	level slog.Level
}

type wsLogHandler struct {
	sync.Mutex
	data   map[*websocket.Conn]chanLevel
	logger slog.Interface
}

func newWSLogHook(logger slog.Interface) *wsLogHandler {
	return &wsLogHandler{
		logger: logger,
		data:   map[*websocket.Conn]chanLevel{},
	}
}

func (h *wsLogHandler) addConn(ws *websocket.Conn, level slog.Level) <-chan struct{} {
	h.Lock()
	defer h.Unlock()

	ch := make(chan struct{})
	h.data[ws] = chanLevel{
		ch:    ch,
		level: level,
	}

	return ch
}

func (h *wsLogHandler) deleteConnNoLock(ws *websocket.Conn) {
	cl, ok := h.data[ws]
	if !ok {
		return
	}

	delete(h.data, ws)

	// let the handler return
	select {
	case cl.ch <- struct{}{}:
	default:
	}
}

type wsMessage struct {
	Fields  map[string]string
	Time    time.Time
	Level   int
	Message string
}

func newWSMessage(entry *slog.Entry) *wsMessage {
	fields := make(map[string]string, len(entry.Fields))
	for key, value := range entry.Fields {
		fields[key] = fmt.Sprintf("%v", value)
	}
	return &wsMessage{
		Fields:  fields,
		Time:    entry.Time,
		Level:   int(entry.Level),
		Message: entry.Message,
	}
}

func (h *wsLogHandler) HandleLog(entry *slog.Entry) error {
	go func() {
		h.Lock()
		defer h.Unlock()

		msg := newWSMessage(entry)

		for ws, cl := range h.data {
			if ws == nil {
				h.deleteConnNoLock(ws)
				continue
			}

			if cl.level < entry.Level {
				continue
			}

			if err := websocket.JSON.Send(ws, msg); err != nil {
				h.deleteConnNoLock(ws)
			}
		}
	}()
	return nil
}
