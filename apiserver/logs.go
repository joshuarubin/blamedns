package apiserver

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type wsLogHook struct {
	sync.Mutex
	conn   map[*websocket.Conn]chan struct{}
	logger *logrus.Logger
}

func newWSLogHook(logger *logrus.Logger) *wsLogHook {
	return &wsLogHook{
		logger: logger,
		conn:   map[*websocket.Conn]chan struct{}{},
	}
}

func (hook *wsLogHook) addConn(ws *websocket.Conn) <-chan struct{} {
	hook.Lock()
	defer hook.Unlock()
	ch := make(chan struct{})
	hook.conn[ws] = ch
	return ch
}

func (hook *wsLogHook) Levels() []logrus.Level {
	ret := make([]logrus.Level, hook.logger.Level+1)

	for i := logrus.PanicLevel; i <= hook.logger.Level; i++ {
		ret[i] = i
	}

	return ret
}

func (hook *wsLogHook) deleteConnNoLock(ws *websocket.Conn) {
	ch, ok := hook.conn[ws]
	if !ok {
		return
	}

	delete(hook.conn, ws)

	select {
	case ch <- struct{}{}:
	default:
	}
}

type wsMessage struct {
	Data    logrus.Fields
	Time    time.Time
	Level   logrus.Level
	Message string
}

func newWSMessage(entry *logrus.Entry) *wsMessage {
	return &wsMessage{
		Data:    entry.Data,
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

		for ws := range hook.conn {
			if ws == nil {
				hook.deleteConnNoLock(ws)
				continue
			}

			if err := websocket.JSON.Send(ws, msg); err != nil {
				hook.deleteConnNoLock(ws)
			}
		}
	}()
	return nil
}
