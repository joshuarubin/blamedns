package simpleserver

import (
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// A Server defines parameters for running a simple http server.
// Addr is required (unless Serve(l) is used directly). Name is recommended.
type Server struct {
	Name    string
	Logger  *logrus.Logger
	Addr    string
	Handler http.Handler
	ln      net.Listener
}

// ServerName just returns the server name. Provided so that structs that stack
// simpleserver will provide this method.
func (s Server) ServerName() string {
	return s.Name
}

// ServerAddr returns the address the server is configured to listen on.
// Provided so that structs that stack simpleserver will provide this method.
func (s Server) ServerAddr() string {
	if s.ln != nil {
		return s.ln.Addr().String()
	}
	return s.Addr
}

// ServeHTTP exists so Server itself satisfies the http.Handler interface. It
// just proxies to Server.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(w, req)
}

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
// If srv.Addr is blank, ":http" is used.
// ListenAndServe always returns a non-nil error.
func (s *Server) ListenAndServe() error {
	if err := s.newListener(); err != nil {
		return err
	}

	return s.Serve(s.ln)
}

func (s *Server) newListener() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.ln = ln
	return nil
}

// Serve accepts incoming connections on the Listener l. It just proxies to
// http.Server.Serve(l).
// Serve always returns a non-nil error.
func (s *Server) Serve(l net.Listener) error {
	s.ln = l

	if s.Logger != nil {
		s.Logger.WithFields(s.logFields()).Infof("starting")
	}

	return (&http.Server{Handler: s.Handler}).Serve(l)
}

// Init prepares the Server for Start. If it does not return an error, then
// Start will not return one either.
func (s *Server) Init() error {
	if s.ln == nil {
		if err := s.newListener(); err != nil {
			return err
		}
	}
	return nil
}

// Start the server in a background goroutine. It only returns initialization
// errors. It will not return an error if Init didn't.
func (s *Server) Start() error {
	if err := s.Init(); err != nil {
		return err
	}

	go func() {
		if err := s.Serve(s.ln); err != nil {
			if nerr, ok := err.(*net.OpError); ok {
				// wish there was a better way to detect this...
				// https://github.com/golang/go/issues/4373
				if nerr.Op == "accept" && nerr.Net == "tcp" && nerr.Err.Error() == "use of closed network connection" {
					// Close sometimes causes this...
					return
				}
			}
			if s.Logger != nil {
				s.Logger.WithError(err).WithFields(s.logFields()).Warnf("serve error")
			}
		}
	}()

	return nil
}

func (s *Server) logFields() logrus.Fields {
	return logrus.Fields{
		"server": s.ServerName(),
		"addr":   s.ServerAddr(),
	}
}

// Close shuts down the server.
func (s *Server) Close() error {
	if s.ln == nil {
		return nil
	}

	if s.Logger != nil {
		s.Logger.WithFields(s.logFields()).Infof("stopping")
	}

	if err := s.ln.Close(); err != nil {
		if s.Logger != nil {
			s.Logger.WithError(err).WithFields(s.logFields()).Warn("error stopping")
		}
		return err
	}

	s.ln = nil
	return nil
}
