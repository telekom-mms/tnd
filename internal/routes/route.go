// Package routes contains components for route watching.
package routes

import (
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Watch waits for routing update events and then probes the
// trusted https servers.
type Watch struct {
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
	// register for route update events
	events := make(chan netlink.RouteUpdate)
	if err := netlink.RouteSubscribe(events, w.done); err != nil {
		log.WithError(err).Fatal("TND route subscribe error")
	}

	// run initial probe
	w.sendProbe()

	// handle route update events
	for e := range events {
		switch e.Type {
		case unix.RTM_NEWROUTE:
			log.WithField("dst", e.Dst).Debug("TND got route NEW event")
		case unix.RTM_DELROUTE:
			log.WithField("dst", e.Dst).Debug("TND got route DEL event")
		}
		w.sendProbe()
	}
}

// Start starts the Watch.
func (w *Watch) Start() {
	go w.start()
}

// Stop stops the Watch.
func (w *Watch) Stop() {
	// NOTE: this will not terminate the Watch goroutine until the
	// next netlink event arrives due to a known issue of the netlink
	// library we use. So, we cannot really wait for the goroutine
	// termination here
	close(w.done)
}

// Probes returns the probe channel.
func (w *Watch) Probes() chan struct{} {
	return w.probes
}

// NewWatch returns a new Watch.
func NewWatch(probes chan struct{}) *Watch {
	return &Watch{
		probes: probes,
		done:   make(chan struct{}),
	}
}
