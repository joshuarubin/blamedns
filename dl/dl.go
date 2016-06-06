package dl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type DL struct {
	URL            *url.URL
	BaseDir        string
	UpdateInterval time.Duration
	Client         Doer
	Logger         *slog.Logger
	fileName       string
	mu             sync.RWMutex
	stopCh         chan struct{}
	AppName        string
	AppVersion     string
	DebugHTTP      bool
}

const (
	DefaultUpdateInterval = 24 * time.Hour
	DefaultAppName        = "DL"
	DefaultAppVersion     = "1.0"
)

func New(u *url.URL, baseDir string) *DL {
	return &DL{
		URL:     u,
		BaseDir: baseDir,
	}
}

func (d *DL) Init() error {
	err := os.MkdirAll(d.BaseDir, 0700)
	if err != nil {
		return err
	}

	file := d.URL.Path

	// strip leading '/'
	if len(file) > 0 && file[0] == '/' {
		file = file[1:]
	}

	// join host and path
	file = strings.Join([]string{d.URL.Host, file}, "__")

	// replace '/' with '__'
	file = strings.Replace(file, "/", "__", -1)

	if d.UpdateInterval == 0 {
		d.UpdateInterval = DefaultUpdateInterval
	}

	if d.Logger == nil {
		d.Logger = text.Logger(slog.InfoLevel)
	}

	if d.Client == nil {
		d.Client = http.DefaultClient
	}

	if len(d.AppName) == 0 {
		d.AppName = DefaultAppName
	}

	if len(d.AppVersion) == 0 {
		d.AppVersion = DefaultAppVersion
	}

	d.fileName = path.Join(d.BaseDir, file)

	return nil
}

func (d *DL) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stopCh != nil {
		return
	}

	d.stopCh = make(chan struct{})

	go func() { _, _ = d.Update() }()

	ticker := time.NewTicker(d.UpdateInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				go func() { _, _ = d.Update() }()
			case <-d.stopCh:
				close(d.stopCh)
				return
			}
		}
	}()
}

func (d *DL) Update() (updated bool, err error) {
	info, err := os.Stat(d.fileName)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	ctxLog := d.Logger.WithFields(slog.Fields{
		"fileName": d.fileName,
		"URL":      d.URL,
		"interval": d.UpdateInterval,
	})

	if !os.IsNotExist(err) {
		if info.IsDir() {
			return false, fmt.Errorf("%s is a directory", d.fileName)
		}

		if info.ModTime().After(time.Now().Add(-d.UpdateInterval)) {
			// file exists and does not need to be updated
			ctxLog.Debug("file does not need to be updated yet")
			return false, nil
		}
	}

	// file doesn't exist or needs to be updated

	ctxLog.Debug("downloading file")

	req, _ := http.NewRequest("GET", d.URL.String(), nil)
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", d.AppName, d.AppVersion))
	d.debugRequestOut(req, false)

	res, err := d.Client.Do(req)
	if err != nil {
		return false, err
	}

	defer func() { _ = res.Body.Close() }()

	d.debugResponse(res, false)

	if res.StatusCode != http.StatusOK {
		err = NewErrStatusCode(res.StatusCode)
		ctxLog.WithError(err).Warn("error updating file")
		return false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// open the file and truncate it if it already exists
	f, err := os.Create(d.fileName)
	if err != nil {
		return false, nil
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		ctxLog.WithError(err).Warn("error updating file")
		return false, err
	}

	ctxLog.Debug("successfully updated file")
	return true, nil
}

func (d *DL) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stopCh == nil {
		return
	}

	d.stopCh <- struct{}{}
	<-d.stopCh
	d.stopCh = nil

	d.Logger.WithFields(slog.Fields{
		"fileName": d.fileName,
		"URL":      d.URL,
	}).Debug("stopped file downloader")
}
