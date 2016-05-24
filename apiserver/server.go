package apiserver

import (
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/Sirupsen/logrus"
)

type Server struct {
	Logger *logrus.Logger
}

var DefaultServer = &Server{Logger: logrus.New()}

func Handler(server *Server) http.Handler {
	ret := http.NewServeMux()

	ret.HandleFunc("/debug/vars", expvarHandler)
	ret.HandleFunc("/debug/pprof/", pprof.Index)
	ret.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ret.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ret.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ret.HandleFunc("/debug/pprof/trace", pprof.Trace)

	ret.Handle("/v1/", v1Handler(server))

	return ret
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: Handler(s)}
	s.Logger.WithField("addr", addr).Info("starting apiserver")
	return srv.ListenAndServe()
}

func (s *Server) Serve(l net.Listener) error {
	srv := &http.Server{Handler: Handler(s)}
	s.Logger.WithField("addr", l.Addr().String()).Info("starting apiserver")
	return srv.Serve(l)
}

func ListenAndServe(addr string) error {
	return DefaultServer.ListenAndServe(addr)
}

func Serve(l net.Listener) error {
	return DefaultServer.Serve(l)
}
