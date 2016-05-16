package hosts

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Hosts struct {
	update   time.Duration
	dir      string
	fileName string
	url      string
}

const subDir = "hosts"

func New(u, cacheDir string, update time.Duration) (*Hosts, error) {
	p, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	dirName := path.Join(cacheDir, subDir, p.Host)
	err = os.MkdirAll(dirName, 0700)
	if err != nil {
		return nil, err
	}

	// strip leading '/'
	if p.Path[0] == '/' {
		p.Path = p.Path[1:]
	}

	// replace '/' with '__'
	p.Path = strings.Replace(p.Path, "/", "__", -1)

	return &Hosts{
		update:   update,
		dir:      dirName,
		fileName: path.Join(dirName, p.Path),
		url:      u,
	}, nil
}

func (h *Hosts) Update() error {
	info, err := os.Stat(h.fileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if !os.IsNotExist(err) {
		if info.IsDir() {
			return fmt.Errorf("%s is a directory", h.fileName)
		}

		if info.ModTime().After(time.Now().Add(-h.update)) {
			// file exists and does not need to be updated
			log.Infof("%s does not need to be updated yet", h.fileName)
			return nil
		}

		// file exists and needs to be updated
	}

	log.Infof("updating %s", h.fileName)

	// open the file and truncate it if it already exists
	f, err := os.Create(h.fileName)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	res, err := http.Get(h.url)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(`\S+\s+(\S+)`)
}

func (h Hosts) Hosts() ([]string, error) {
	f, err := os.Open(h.fileName)
	if err != nil {
		return nil, err
	}

	var ret []string

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

		ret = append(ret, host)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}
