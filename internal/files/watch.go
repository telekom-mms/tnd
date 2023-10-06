// Package files contains components for file watching.
package files

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

const (
	// resolv.conf files in /etc and /run/systemd/resolve.
	etc               = "/etc"
	etcResolvConf     = etc + "/resolv.conf"
	systemdResolveDir = "/run/systemd/resolve"
	systemdResolvConf = systemdResolveDir + "/resolv.conf"
	stubResolvConf    = systemdResolveDir + "/stub-resolv.conf"
)

// Watch watches resolv.conf files and then probes the trusted https servers.
type Watch struct {
	probes chan struct{}
	done   chan struct{}
	closed chan struct{}
}

// sendProbe sends a probe request over the probe channel.
func (w *Watch) sendProbe() {
	select {
	case w.probes <- struct{}{}:
	case <-w.done:
	}
}

// isResolvConfEvent checks if event is a resolv.conf file event.
func isResolvConfEvent(event fsnotify.Event) bool {
	switch event.Name {
	case etcResolvConf:
		return true
	case stubResolvConf:
		return true
	case systemdResolvConf:
		return true
	}
	return false
}

// start starts the Watch.
func (w *Watch) start() {
	defer close(w.closed)

	// create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Fatal("TND could not create file watcher")
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.WithError(err).Error("TND could not stop file watcher")
		}
	}()

	// add resolv.conf folders to watcher
	if err := watcher.Add(etc); err != nil {
		log.WithError(err).Debug("TND could not add etc to file watcher")
	}
	if err := watcher.Add(systemdResolveDir); err != nil {
		log.WithError(err).Debug("TND could not add systemd to file watcher")
	}

	// run initial probe
	w.sendProbe()

	// watch the files
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if isResolvConfEvent(event) {
				log.WithFields(log.Fields{
					"name": event.Name,
					"op":   event.Op,
				}).Debug("TND got resolv.conf file event")
				w.sendProbe()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.WithError(err).Error("TND got error file event")
		case <-w.done:
			return
		}
	}
}

// Start starts the Watch.
func (w *Watch) Start() {
	go w.start()
}

// Stop stops the Watch.
func (w *Watch) Stop() {
	close(w.done)
	<-w.closed
}

// NewWatch returns a new Watch.
func NewWatch(probes chan struct{}) *Watch {
	return &Watch{
		probes: probes,
		done:   make(chan struct{}),
		closed: make(chan struct{}),
	}
}
