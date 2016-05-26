package pixelserv

import (
	"net"
	"net/http"

	"jrubin.io/blamedns/simpleserver"

	"github.com/Sirupsen/logrus"
)

// PixelGIF is a single pixel transparent GIF image.
var PixelGIF = []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")

// DefaultAddr is used by Start and ListenAndServe if no Addr had been provided.
const DefaultAddr = ":http"

// A Server defines parameters for running a pixelserv http server.
type Server struct {
	simpleserver.Server
}

const name = "pixelserv"

// New allocates a new pixelserv Server.
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

// DefaultServer is the default Server used by Serve.
var DefaultServer = New(DefaultAddr, logrus.New())

// Handler returns an http.Handler that will simply respond with a pixel for any
// request.
func Handler(logger *logrus.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		if _, err := w.Write(PixelGIF); err != nil && logger != nil {
			logger.WithError(err).Warn("pixelserv handler error")
		}
	})
}

// ListenAndServe listens on the TCP network address addr
// and then calls Serve.
func ListenAndServe(addr string) error {
	DefaultServer.Server.Addr = addr
	return DefaultServer.ListenAndServe()
}

// Serve uses the DefaultServer and accepts incoming connections on the Listener
// l. The service responds with the pixel for every request. Serve always
// returns a non-nil error. It is the caller's responsibility to call l.Close(),
// or Server.Stop() to shutdown the server.
func Serve(l net.Listener) error {
	return DefaultServer.Serve(l)
}

// Close shuts down the DefaultServer server.
func Close() error {
	return DefaultServer.Close()
}

// Start the DefaultServer in a background goroutine. It only returns
// initialization errors. It will not return an error if Init didn't.
func Start() error {
	return DefaultServer.Start()
}
