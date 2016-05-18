package pixelserv

import (
	"net"
	"net/http"
)

var pixelGIF = []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")

func ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: Handler}
	return srv.ListenAndServe()
}

func Serve(l net.Listener) error {
	srv := &http.Server{Handler: Handler}
	return srv.Serve(l)
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/gif")
	w.Write(pixelGIF)
})
