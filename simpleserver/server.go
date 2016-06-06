package simpleserver

import (
	"log"
	"net"
	"net/http"
	"time"

	"jrubin.io/slog"

	"gopkg.in/tylerb/graceful.v1"
)

// A Server defines parameters for running a simple http server.
// Addr is required (unless Serve(l) is used directly). Name is recommended.
type Server struct {
	Name        string
	Logger      *slog.Logger
	Addr        string
	Handler     http.Handler
	StopTimeout time.Duration
	ln          net.Listener
	srv         *graceful.Server
}

const DefaultStopTimeout = 10 * time.Second

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
	s.logHandler().ServeHTTP(w, req)
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

	var debugLogger, errorLogger *log.Logger

	if s.Logger != nil {
		s.Logger.WithFields(slog.Fields{
			"server": s.ServerName(),
			"addr":   s.ServerAddr(),
		}).Info("starting")

		w := s.Logger.Writer(slog.DebugLevel)
		debugLogger = log.New(w, "", 0)

		w = s.Logger.Writer(slog.ErrorLevel)
		errorLogger = log.New(w, "", 0)
	}

	s.srv = &graceful.Server{
		Server: &http.Server{
			Handler:  s.logHandler(),
			ErrorLog: errorLogger,
		},
		NoSignalHandling: true,
		Logger:           debugLogger,
	}

	return s.srv.Serve(l)
}

func (s Server) stopTimeout() time.Duration {
	if s.StopTimeout > 0 {
		return s.StopTimeout
	}
	return DefaultStopTimeout
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
		if err := s.Serve(s.ln); err != nil && s.Logger != nil {
			s.Logger.WithError(err).WithFields(slog.Fields{
				"server": s.ServerName(),
				"addr":   s.ServerAddr(),
			}).Warn("serve error")
		}
	}()

	return nil
}

// Close shuts down the server.
func (s *Server) Close() error {
	if s.srv == nil {
		return nil
	}

	if s.Logger != nil {
		s.Logger.WithFields(slog.Fields{
			"server": s.ServerName(),
			"addr":   s.ServerAddr(),
		}).Info("stopping")
	}

	ch := s.srv.StopChan()
	s.srv.Stop(s.stopTimeout())
	<-ch

	s.srv = nil
	s.ln = nil

	return nil
}
