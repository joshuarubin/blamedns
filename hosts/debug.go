package hosts

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/Sirupsen/logrus"
)

type debugHTTPFn func(interface{}, bool) ([]byte, error)

func (h *Hosts) debugRequestOut(req *http.Request, b bool) {
	h.debugHTTP(req, dumpRequestOut, "> ", b)
}

func (h *Hosts) debugResponse(resp *http.Response, b bool) {
	h.debugHTTP(resp, dumpResponse, "< ", b)
}

func dumpRequestOut(req interface{}, b bool) ([]byte, error) {
	return httputil.DumpRequestOut(req.(*http.Request), b)
}

func dumpResponse(resp interface{}, b bool) ([]byte, error) {
	return httputil.DumpResponse(resp.(*http.Response), b)
}

func (h *Hosts) debugHTTP(data interface{}, fn debugHTTPFn, prefix string, b bool) {
	dump, err := fn(data, b)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	debugWriter := h.Logger.WriterLevel(logrus.DebugLevel)
	defer func() { _ = debugWriter.Close() }()

	for _, line := range strings.Split(string(dump), "\n") {
		fmt.Fprintf(debugWriter, "%s%s\n", prefix, line)
	}
}
