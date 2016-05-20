package watcher

import (
	"bufio"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"jrubin.io/blamedns/parser"
)

type Watcher struct {
	Dir        []string
	Parser     parser.Parser
	Logger     *logrus.Logger
	watcher    *fsnotify.Watcher
	stopCh     chan struct{}
	parseTimer map[string]*time.Timer
	mu         sync.Mutex
}

func New(l *logrus.Logger, p parser.Parser, dir ...string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, d := range dir {
		if err = w.Add(d); err != nil {
			return nil, err
		}
	}

	if l == nil {
		l = logrus.New()
	}

	l.WithField("directories", strings.Join(dir, ", ")).Debug("watching directories")

	return &Watcher{
		Dir:        dir,
		Parser:     p,
		Logger:     l,
		watcher:    w,
		parseTimer: map[string]*time.Timer{},
	}, nil
}

func (w *Watcher) parse(file string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.parseTimer, file)

	f, err := os.Open(file)
	if err != nil {
		w.Logger.WithError(err).WithField("file", file).Warn("error opening file")
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	i := 0
	n := 0
	for scanner.Scan() {
		i++
		if w.Parser.Parse(file, i, scanner.Text()) {
			n++
		}
	}

	if err := scanner.Err(); err != nil {
		w.Logger.WithError(err).WithField("file", file).Warn("error scanning file")
		return
	}

	w.Logger.WithFields(logrus.Fields{
		"file": file,
		"num":  n,
	}).Debug("parsed")
}

// this much time must elapse without a file being further modified before it
// will be parsed
const debounce = 500 * time.Millisecond

func (w *Watcher) setParseTimer(file string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, ok := w.parseTimer[file]; ok {
		t.Reset(debounce)
		return
	}

	w.parseTimer[file] = time.AfterFunc(debounce, func() {
		w.parse(file)
	})
}

func (w *Watcher) parseAll() {
	for _, d := range w.Dir {
		f, err := os.Open(d)
		if err != nil {
			w.Logger.WithError(err).WithField("dir", d).Warn("error opening dir")
			continue
		}

		defer func() { _ = f.Close() }()

		fi, err := f.Readdir(-1)
		if err != nil {
			w.Logger.WithError(err).WithField("dir", d).Warn("error reading dir")
			continue
		}

		for _, fd := range fi {
			w.parse(path.Join(d, fd.Name()))
		}
	}
}

func (w *Watcher) Start() {
	if w.stopCh != nil {
		return
	}

	w.stopCh = make(chan struct{})

	// make sure files are parsed when firsted started
	w.parseAll()

	go func() {
		for {
			select {
			case event := <-w.watcher.Events:
				if event.Op&fsnotify.Chmod > 0 {
					w.parse(event.Name)
					break
				}

				if event.Op&(fsnotify.Write|fsnotify.Create) > 0 {
					w.setParseTimer(event.Name)
				}
			case err := <-w.watcher.Errors:
				w.Logger.WithError(err).Warn("watcher error")
			case <-w.stopCh:
				// TODO(jrubin) should we do this?
				// defer w.Close()

				close(w.stopCh)
				return
			}
		}
	}()
}

func (w *Watcher) Stop() {
	if w.stopCh == nil {
		return
	}

	w.stopCh <- struct{}{}
	<-w.stopCh
	w.stopCh = nil

	w.Logger.WithField("directories", strings.Join(w.Dir, ", ")).Debug("stopped watching directories")
}
