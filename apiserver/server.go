package apiserver

import (
	"net/http"
	"net/http/pprof"

	"jrubin.io/blamedns/simpleserver"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

// A Server defines parameters for running an apiserver http server.
type Server struct {
	simpleserver.Server
}

const name = "apiserver"

// New allocates a new apiserver Server.
func New(addr string, logger *logrus.Logger) *Server {
	return &Server{
		Server: simpleserver.Server{
			Name:    name,
			Logger:  logger,
			Addr:    addr,
			Handler: Handler(logger),
		},
	}
}

// Handler returns an http.Handler that responds properly to all apiserver
// routes.
func Handler(logger *logrus.Logger) http.Handler {
	ret := http.NewServeMux()

	ret.HandleFunc("/debug/vars", expvarHandler)
	ret.HandleFunc("/debug/pprof/", pprof.Index)
	ret.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ret.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ret.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ret.HandleFunc("/debug/pprof/trace", pprof.Trace)

	ret.Handle("/metrics", prometheus.Handler())

	ret.Handle("/v1/", v1Handler(logger, "/v1"))

	return ret
}
