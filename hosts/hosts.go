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

	log "github.com/Sirupsen/logrus"
)

type Hosts struct {
	mu          sync.RWMutex
	update      time.Duration
	dir         string
	fileName    string
	url         string
	stopCh      chan struct{}
	hosts       map[string]struct{}
	initialized bool
}

const subDir = "hosts"

func New(u, cacheDir string, update time.Duration) (*Hosts, error) {
	p, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	dirName := path.Join(cacheDir, subDir)
	err = os.MkdirAll(dirName, 0700)
	if err != nil {
		return nil, err
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

	return &Hosts{
		update:   update,
		dir:      dirName,
		fileName: path.Join(dirName, file),
		url:      u,
	}, nil
}

func (h *Hosts) Start() (<-chan struct{}, <-chan error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopCh != nil {
		return nil, nil
	}

	h.stopCh = make(chan struct{})

	log.Infof("starting hosts file %s", h.fileName)

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
			case <-time.Tick(h.update):
				go updater()
			case <-h.stopCh:
				log.Infof("stopped hosts file %s", h.fileName)
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

	log.Infof("stopping hosts file %s", h.fileName)
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
	h.mu.Lock()
	defer h.mu.Unlock()

	info, err := os.Stat(h.fileName)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if !os.IsNotExist(err) {
		if info.IsDir() {
			return false, fmt.Errorf("%s is a directory", h.fileName)
		}

		if info.ModTime().After(time.Now().Add(-h.update)) {
			// file exists and does not need to be updated
			log.Infof("%s does not need to be updated yet", h.fileName)
			return false, nil
		}

		// file exists and needs to be updated
	}

	log.Infof("updating %s", h.fileName)

	res, err := http.Get(h.url)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	// open the file and truncate it if it already exists
	f, err := os.Create(h.fileName)
	if err != nil {
		return false, nil
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		log.Warnf("error updating %s: %v", h.fileName, err)
		return false, err
	}

	log.Infof("successfully updated %s", h.fileName)
	return true, nil
}

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(`\S+\s+(\S+)`)
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

	log.Debugf("parsing %s", h.fileName)

	f, err := os.Open(h.fileName)
	if err != nil {
		return err
	}

	h.hosts = map[string]struct{}{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		i := strings.Index(text, "#")
		if i == 0 {
			continue
		}
		if i != -1 {
			text = text[:i-1]
		}

		text = strings.TrimSpace(text)
		if len(text) == 0 {
			continue
		}

		res := re.FindStringSubmatch(text)
		if len(res) < 2 {
			continue
		}

		host := res[1]
		if host == "localhost" {
			continue
		}

		if host == "localhost.localdomain" {
			continue
		}

		if host == "broadcasthost" {
			continue
		}

		if host == "local" {
			continue
		}

		h.hosts[host+"."] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	log.Infof("parsed %d hosts in %s", len(h.hosts), h.fileName)

	h.initialized = true
	return nil
}
