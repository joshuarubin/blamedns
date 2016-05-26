package apiserver

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

func v1Handler(logger *logrus.Logger) http.Handler {
	ret := http.NewServeMux()
	ret.HandleFunc("/v1/ping", v1PingHandler(logger))
	return ret
}

func v1PingHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("pong")
		fmt.Fprintf(w, "pong\n")
	}
}
