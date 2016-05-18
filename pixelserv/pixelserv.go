package pixelserv

import (
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
)

var pixelGIF = []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")

type Server struct {
	Logger *logrus.Logger
}

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/gif")
	if _, err := w.Write(pixelGIF); err != nil && s.Logger != nil {
		s.Logger.WithError(err).Warn("pixelserv handler error")
	}
}

func (s Server) logStart(addr string) {
	if s.Logger == nil {
		return
	}

	s.Logger.WithFields(logrus.Fields{
		"addr": addr,
	}).Info("starting pixelserv")
}

func (s Server) ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: s}
	s.logStart(addr)
	return srv.ListenAndServe()
}

func (s Server) Serve(l net.Listener) error {
	srv := &http.Server{Handler: s}
	s.logStart(l.Addr().String())
	return srv.Serve(l)
}

var DefaultServer = &Server{Logger: logrus.New()}

func ListenAndServe(addr string) error {
	return DefaultServer.ListenAndServe(addr)
}

func Serve(l net.Listener) error {
	return DefaultServer.Serve(l)
}

var Handler = DefaultServer.ServeHTTP
