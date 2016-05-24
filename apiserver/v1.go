package apiserver

import (
	"fmt"
	"net/http"
)

func v1Handler(server *Server) http.Handler {
	ret := http.NewServeMux()
	ret.HandleFunc("/v1/ping", v1PingHandler(server))
	return ret
}

func v1PingHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.Logger.Debug("pong")
		fmt.Fprintf(w, "pong\n")
	}
}
