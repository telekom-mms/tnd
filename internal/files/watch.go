// Package files contains components for file watching.
package files

import (
	"path/filepath"
	"slices"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

// Watcher is the file watcher interface.
type Watcher interface {
	Start() error
	Stop()
}

// Watch watches resolv.conf files and then probes the trusted https servers.
type Watch struct {
	files   []string
	watcher *fsnotify.Watcher
	probes  chan struct{}
	done    chan struct{}
	closed  chan struct{}
}

// sendProbe sends a probe request over the probe channel.
func (w *Watch) sendProbe() {
	select {
	case w.probes <- struct{}{}:
	case <-w.done:
	}
}

// start starts the Watch.
func (w *Watch) start() {
	defer close(w.closed)
	defer func() {
		if err := w.watcher.Close(); err != nil {
			log.WithError(err).Error("TND could not stop file watcher")
		}
	}()

	// run initial probe
	w.sendProbe()

	// watch the files
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if slices.Contains(w.files, event.Name) {
				log.WithFields(log.Fields{
					"name": event.Name,
					"op":   event.Op,
				}).Debug("TND got resolv.conf file event")
				w.sendProbe()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.WithError(err).Error("TND got error file event")
		case <-w.done:
			return
		}
	}
}

// fsnotifyNewWatcher is fsnotify.NewWatcher for testing.
var fsnotifyNewWatcher = fsnotify.NewWatcher

// watcherAdd is watcher.Add for testing.
var watcherAdd = func(watcher *fsnotify.Watcher, name string) error {
	return watcher.Add(name)
}

// Start starts the Watch.
func (w *Watch) Start() error {
	// create watcher
	watcher, err := fsnotifyNewWatcher()
	if err != nil {
		log.WithError(err).Error("TND could not create file watcher")
		return err
	}

	// add resolv.conf folders to watcher
	for _, f := range w.files {
		p := filepath.Dir(f)
		if err := watcherAdd(watcher, p); err != nil {
			log.WithError(err).
				WithField("file", f).
				WithField("folder", p).
				Error("TND could not add folder to file watcher")
			_ = watcher.Close()
			return err
		}
	}

	// start watcher
	w.watcher = watcher
	go w.start()
	return nil
}

// Stop stops the Watch.
func (w *Watch) Stop() {
	close(w.done)
	<-w.closed
}

// NewWatch returns a new Watch.
func NewWatch(probes chan struct{}, files []string) *Watch {
	return &Watch{
		files:  files,
		probes: probes,
		done:   make(chan struct{}),
		closed: make(chan struct{}),
	}
}
