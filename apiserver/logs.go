package apiserver

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

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
