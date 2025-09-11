// Package routes contains components for route watching.
package routes

import (
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Watcher is the watcher interface.
type Watcher interface {
	Start() error
	Stop()
}

// Watch waits for routing update events and then probes the
// trusted https servers.
type Watch struct {
	events chan netlink.RouteUpdate
	probes chan struct{}
	done   chan struct{}
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
	// run initial probe
	w.sendProbe()

	// handle route update events
	for e := range w.events {
		switch e.Type {
		case unix.RTM_NEWROUTE:
			log.WithField("dst", e.Dst).Debug("TND got route NEW event")
		case unix.RTM_DELROUTE:
			log.WithField("dst", e.Dst).Debug("TND got route DEL event")
		}
		w.sendProbe()
	}
}

// netlinkRouteSubscribe is netlink.RouteSubscribe for testing.
var netlinkRouteSubscribe = netlink.RouteSubscribe

// Start starts the Watch.
func (w *Watch) Start() error {
	// register for route update events
	if err := netlinkRouteSubscribe(w.events, w.done); err != nil {
		log.WithError(err).Error("TND route subscribe error")
		return err
	}

	// start watcher
	go w.start()
	return nil
}

// Stop stops the Watch.
func (w *Watch) Stop() {
	// NOTE: this will not terminate the Watch goroutine until the
	// next netlink event arrives due to a known issue of the netlink
	// library we use. So, we cannot really wait for the goroutine
	// termination here
	close(w.done)
}

// NewWatch returns a new Watch.
func NewWatch(probes chan struct{}) *Watch {
	return &Watch{
		events: make(chan netlink.RouteUpdate),
		probes: probes,
		done:   make(chan struct{}),
	}
}
