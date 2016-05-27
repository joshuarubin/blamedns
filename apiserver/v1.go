package apiserver

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"

	"github.com/Sirupsen/logrus"
)

func v1Handler(logger *logrus.Logger) http.Handler {
	ret := http.NewServeMux()
	ret.HandleFunc("/v1/ping", v1PingHandler)
	ret.Handle("/v1/logs", v1Logs(logger))
	return ret
}

func v1PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func v1Logs(logger *logrus.Logger) websocket.Handler {
	wsLogger := newWSLogHook(logger)
	logger.Hooks.Add(wsLogger)

	return func(ws *websocket.Conn) {
		<-wsLogger.addConn(ws)
	}
}
