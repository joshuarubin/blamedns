package apiserver

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// A Leveler is used to determine from an http.Request the logLevel that should
// be used for the websocket.
type Leveler interface {
	Level(req *http.Request) logrus.Level
}

// LevelerFunc is a convenience wrapper to create a Leveler.
type LevelerFunc func(*http.Request) logrus.Level

// Level implements the Leveler interface.
func (f LevelerFunc) Level(req *http.Request) logrus.Level {
	return f(req)
}

// LogsHandler returns an http.Handler that will push JSON formatted log
// messages to the client. Handler should only be called once per logger. By
// default, each request is configured to receive logs at logger.Level or
// higher. If you would like to be able to set a custom logLevel per websocket
// connection, use the Leveler interface to determine what level the request
// should use.
func LogsHandler(logger *logrus.Logger, leveler Leveler) http.Handler {
	wsLogger := newWSLogHook(logger)
	logger.Hooks.Add(wsLogger)

	return websocket.Handler(func(ws *websocket.Conn) {
		level := logger.Level
		if leveler != nil {
			level = leveler.Level(ws.Request())
		}

		<-wsLogger.addConn(ws, level)
	})
}

type chanLevel struct {
	ch    chan struct{}
	level logrus.Level
}

type wsLogHook struct {
	sync.Mutex
	data   map[*websocket.Conn]chanLevel
	logger *logrus.Logger
}

func newWSLogHook(logger *logrus.Logger) *wsLogHook {
	return &wsLogHook{
		logger: logger,
		data:   map[*websocket.Conn]chanLevel{},
	}
}

func (hook *wsLogHook) addConn(ws *websocket.Conn, level logrus.Level) <-chan struct{} {
	hook.Lock()
	defer hook.Unlock()

	ch := make(chan struct{})
	hook.data[ws] = chanLevel{
		ch:    ch,
		level: level,
	}

	return ch
}

func (hook *wsLogHook) Levels() []logrus.Level {
	ret := make([]logrus.Level, logrus.DebugLevel+1)

	// add all levels
	for i := logrus.PanicLevel; i <= logrus.DebugLevel; i++ {
		ret[i] = i
	}

	return ret
}

func (hook *wsLogHook) deleteConnNoLock(ws *websocket.Conn) {
	cl, ok := hook.data[ws]
	if !ok {
		return
	}

	delete(hook.data, ws)

	// let the handler return
	select {
	case cl.ch <- struct{}{}:
	default:
	}
}

type wsMessage struct {
	Data    map[string]string
	Time    time.Time
	Level   logrus.Level
	Message string
}

func newWSMessage(entry *logrus.Entry) *wsMessage {
	data := make(map[string]string, len(entry.Data))
	for key, value := range entry.Data {
		data[key] = fmt.Sprintf("%v", value)
	}
	return &wsMessage{
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}
}

func (hook *wsLogHook) Fire(entry *logrus.Entry) error {
	go func() {
		hook.Lock()
		defer hook.Unlock()

		msg := newWSMessage(entry)

		for ws, cl := range hook.data {
			if ws == nil {
				hook.deleteConnNoLock(ws)
				continue
			}

			if cl.level < entry.Level {
				continue
			}

			if err := websocket.JSON.Send(ws, msg); err != nil {
				hook.deleteConnNoLock(ws)
			}
		}
	}()
	return nil
}
