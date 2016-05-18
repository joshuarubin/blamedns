package hosts

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Hosts struct {
	mu             sync.RWMutex
	dir            string
	fileName       string
	url            string
	stopCh         chan struct{}
	hosts          map[string]struct{}
	initialized    bool
	UpdateInterval time.Duration
	Parser         Parser
	Client         Doer
	Logger         *logrus.Logger
	AppName        string
	AppVersion     string
}

const (
	defaultUpdateInterval = 24 * time.Hour
	subDir                = "hosts"
)

func (h *Hosts) Init(u, cacheDir string) error {
	p, err := url.Parse(u)
	if err != nil {
		return err
	}

	dirName := path.Join(cacheDir, subDir)
	err = os.MkdirAll(dirName, 0700)
	if err != nil {
		return err
	}

	file := p.Path

	// strip leading '/'
	if len(file) > 0 && file[0] == '/' {
		file = file[1:]
	}

	// join host and path
	file = strings.Join([]string{p.Host, file}, "__")

	// replace '/' with '__'
	file = strings.Replace(file, "/", "__", -1)

	h.dir = dirName
	h.fileName = path.Join(dirName, file)
	h.url = u

	if h.UpdateInterval == 0 {
		h.UpdateInterval = defaultUpdateInterval
	}

	if h.Parser == nil {
		h.Parser = FileParser
	}

	if h.Logger == nil {
		h.Logger = logrus.New()
	}

	if h.Client == nil {
		h.Client = http.DefaultClient
	}

	return nil
}

func (h *Hosts) Start() (<-chan struct{}, <-chan error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopCh != nil {
		return nil, nil
	}

	h.stopCh = make(chan struct{})

	h.Logger.Debugf("starting hosts file %s", h.fileName)

	updatedCh := make(chan struct{})
	errCh := make(chan error)

	updater := func() {
		updated, err := h.Update()
		if err != nil {
			errCh <- err
		} else if updated {
			updatedCh <- struct{}{}
		}
	}

	go updater()

	go func() {
		for {
			select {
			case <-time.Tick(h.UpdateInterval):
				go updater()
			case <-h.stopCh:
				h.Logger.Debugf("stopped hosts file %s", h.fileName)
				close(h.stopCh)
				return
			}
		}
	}()

	return updatedCh, errCh
}

func (h *Hosts) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopCh == nil {
		return
	}

	h.Logger.Debugf("stopping hosts file %s", h.fileName)
	h.stopCh <- struct{}{}
	<-h.stopCh
	h.stopCh = nil
}

func (h *Hosts) Update() (bool, error) {
	updated, err := h.doUpdate()
	if err != nil {
		return false, err
	}

	h.mu.RLock()
	initialized := h.initialized
	h.mu.RUnlock()

	if !updated && initialized {
		return false, nil
	}

	if err := h.parse(); err != nil {
		return false, err
	}

	return true, nil
}

func (h *Hosts) doUpdate() (updated bool, err error) {
	info, err := os.Stat(h.fileName)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if !os.IsNotExist(err) {
		if info.IsDir() {
			return false, fmt.Errorf("%s is a directory", h.fileName)
		}

		if info.ModTime().After(time.Now().Add(-h.UpdateInterval)) {
			// file exists and does not need to be updated
			h.Logger.Debugf("%s does not need to be updated yet", h.fileName)
			return false, nil
		}

		// file exists and needs to be updated
	}

	h.Logger.Infof("updating %s", h.fileName)

	req, _ := http.NewRequest("GET", h.url, nil)
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", h.AppName, h.AppVersion))
	h.debugRequestOut(req, false)

	res, err := h.Client.Do(req)
	if err != nil {
		return false, err
	}

	defer func() { _ = res.Body.Close() }()

	h.debugResponse(res, false)

	if res.StatusCode != http.StatusOK {
		err = NewErrStatusCode(res.StatusCode)
		h.Logger.
			WithError(err).
			WithFields(logrus.Fields{
				"fileName": h.fileName,
				"url":      h.url,
			}).
			Warn("error updating hosts")
		return false, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// open the file and truncate it if it already exists
	f, err := os.Create(h.fileName)
	if err != nil {
		return false, nil
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		h.Logger.Warnf("error updating %s: %v", h.fileName, err)
		return false, err
	}

	h.Logger.Debugf("successfully updated %s", h.fileName)
	return true, nil
}

func (h *Hosts) Block(host string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, found := h.hosts[host]
	return found
}

func (h *Hosts) parse() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Logger.Debugf("parsing %s", h.fileName)

	f, err := os.Open(h.fileName)
	if err != nil {
		return err
	}

	h.hosts = map[string]struct{}{}

	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		i++

		text := scanner.Text()
		text = stripComments(text)

		if len(text) == 0 {
			continue
		}

		host, valid := h.Parser.Parse(text)
		if !valid {
			continue
		}

		host = strings.ToLower(host)
		host = stripTrailing(host, ".")
		host = stripPort(host)

		if h.validateHost(i, host) {
			h.hosts[host] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	h.Logger.Debugf("parsed %d hosts in %s", len(h.hosts), h.fileName)

	h.initialized = true
	return nil
}

func stripComments(text string) string {
	i := strings.Index(text, "#")
	if i == 0 {
		return ""
	}

	if i != -1 {
		return text[:i-1]
	}

	return text
}

var portRe = regexp.MustCompile(`(\S+):\d+`)

func stripPort(text string) string {
	res := portRe.FindStringSubmatch(text)
	if len(res) < 2 {
		return text
	}
	return res[1]
}

func stripTrailing(text string, trail string) string {
	i := strings.LastIndex(text, trail)
	if i == -1 {
		return text
	}

	if len(text) != i+len(trail) {
		return text
	}

	return text[:i]
}

func (h *Hosts) validateHost(ln int, host string) bool {
	// the entire hostname (including the delimiting dots but not a trailing
	// dot) has a maximum of 253 ascii characters
	if len(host) > 253 {
		h.Logger.Warnf("%s:%d: hostname too long: %s", h.fileName, ln, host)
		return false
	}

	labels := strings.Split(host, ".")
	for _, label := range labels {
		// each label must be between 1 and 63 characters long
		l := len(label)

		if l == 0 {
			h.Logger.Warnf("%s:%d: empty label: %s", h.fileName, ln, host)
			return false
		}

		if l > 63 {
			h.Logger.Warnf("%s:%d: label too long: %s (%s)", h.fileName, ln, label, host)
			return false
		}

		// hostname labels may contain only the ASCII letters 'a' through 'z' (in a
		// case-insensitive manner), the digits '0' through '9', and the hyphen
		// ('-')
		for i := 0; i < l; i++ {
			c := label[i]

			if c >= 'a' && c <= 'z' {
				continue
			}

			// if c >= 'A' && c <= 'Z' {
			// 	continue
			// }

			if c >= '0' && c <= '9' {
				continue
			}

			if c == '_' {
				continue
			}

			// must not start or end with a hyphen
			if c == '-' && i > 0 && i < l-1 {
				continue
			}

			h.Logger.Warnf("%s:%d: invalid character \"%c\" (%s)", h.fileName, ln, c, host)
			return false
		}
	}

	return true
}
