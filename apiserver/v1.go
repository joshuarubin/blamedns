package apiserver

import (
	"fmt"
	"net/http"

	"jrubin.io/blamedns/bdconfig/bdtype"

	"golang.org/x/net/websocket"

	"github.com/Sirupsen/logrus"
	"github.com/go-zoo/bone"
)

func v1Handler(logger *logrus.Logger, prefix string) http.Handler {
	ret := bone.New().Prefix(prefix)
	ret.GetFunc("/ping", v1PingHandler)

	logsHandler := v1Logs(logger)

	ret.Get("/logs", logsHandler)
	ret.Get("/logs/:level", logsHandler)

	return ret
}

func v1PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func v1Logs(logger *logrus.Logger) websocket.Handler {
	wsLogger := newWSLogHook(logger)
	logger.Hooks.Add(wsLogger)

	return func(ws *websocket.Conn) {
		level := logger.Level
		if val := bone.GetValue(ws.Request(), "level"); len(val) > 0 {
			level = bdtype.ParseLogLevel(val).Level()
		}

		<-wsLogger.addConn(ws, level)
	}
}
