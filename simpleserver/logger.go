package simpleserver

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"jrubin.io/slog"
)

type wrappedWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrappedWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *wrappedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func clientIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func (s Server) logHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()

		ww := &wrappedWriter{ResponseWriter: w, status: 200}
		s.Handler.ServeHTTP(ww, r)

		s.Logger.WithFields(slog.Fields{
			"status":     ww.status,
			"method":     r.Method,
			"host":       r.Host,
			"path":       r.URL.Path,
			"ip":         clientIP(r),
			"latency":    time.Since(begin),
			"user-agent": r.UserAgent(),
			"server":     s.ServerName(),
		}).Debug("http request")
	})
}
