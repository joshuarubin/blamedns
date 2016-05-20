package dl

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/Sirupsen/logrus"
)

type debugHTTPFn func(interface{}, bool) ([]byte, error)

func (d *DL) debugRequestOut(req *http.Request, b bool) {
	d.debugHTTP(req, dumpRequestOut, "> ", b)
}

func (d *DL) debugResponse(resp *http.Response, b bool) {
	d.debugHTTP(resp, dumpResponse, "< ", b)
}

func dumpRequestOut(req interface{}, b bool) ([]byte, error) {
	return httputil.DumpRequestOut(req.(*http.Request), b)
}

func dumpResponse(resp interface{}, b bool) ([]byte, error) {
	return httputil.DumpResponse(resp.(*http.Response), b)
}

func (d *DL) debugHTTP(data interface{}, fn debugHTTPFn, prefix string, b bool) {
	if !d.DebugHTTP {
		return
	}

	dump, err := fn(data, b)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	debugWriter := d.Logger.WriterLevel(logrus.DebugLevel)
	defer func() { _ = debugWriter.Close() }()

	for _, line := range strings.Split(string(dump), "\n") {
		fmt.Fprintf(debugWriter, "%s%s\n", prefix, line)
	}
}
