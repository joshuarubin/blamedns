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
	ret.Get("/logs/:level", v1Logs(logger))

	return ret
}

func v1PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func v1Logs(logger *logrus.Logger) websocket.Handler {
	wsLogger := newWSLogHook(logger)
	logger.Hooks.Add(wsLogger)

	return func(ws *websocket.Conn) {
		val := bone.GetValue(ws.Request(), "level")
		if len(val) == 0 {
			panic(fmt.Errorf("could not determine log level"))
		}
		level := bdtype.ParseLogLevel(val).Level()

		<-wsLogger.addConn(ws, level)
	}
}
