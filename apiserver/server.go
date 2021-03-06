package apiserver

import (
	"net/http"
	"net/http/pprof"

	"jrubin.io/blamedns/simpleserver"
	"jrubin.io/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// A Server defines parameters for running an apiserver http server.
type Server struct {
	simpleserver.Server
}

const name = "apiserver"

// New allocates a new apiserver Server.
func New(addr string, logger *slog.Logger, defaultLogLevel slog.Level) *Server {
	return &Server{
		Server: simpleserver.Server{
			Name:    name,
			Logger:  logger,
			Addr:    addr,
			Handler: Handler(logger, defaultLogLevel),
		},
	}
}

// Handler returns an http.Handler that responds properly to all apiserver
// routes.
func Handler(logger *slog.Logger, defaultLogLevel slog.Level) http.Handler {
	ret := http.NewServeMux()

	ret.HandleFunc("/debug/pprof/", pprof.Index)
	ret.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ret.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ret.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ret.HandleFunc("/debug/pprof/trace", pprof.Trace)

	ret.Handle("/metrics", prometheus.Handler())

	ret.Handle("/logs/", LogsHandler(logger, defaultLogLevel, reqLeveler("/logs/", defaultLogLevel)))

	ret.Handle("/ui/", uiHandler("/ui"))
	ret.Handle("/", http.RedirectHandler("/ui/", http.StatusFound))

	return ret
}

func reqLeveler(prefix string, defaultLogLevel slog.Level) LevelerFunc {
	return func(req *http.Request) slog.Level {
		return slog.ParseLevel(req.URL.Path[len(prefix):], defaultLogLevel)
	}
}

func uiHandler(prefix string) http.Handler {
	fs := http.FileServer(assetFS())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len(prefix):]
		if _, ok := _bindata["public"+r.URL.Path]; !ok {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	})
}

//go:generate go-bindata-assetfs -nometadata -pkg apiserver -prefix ../ui ../ui/public/...
