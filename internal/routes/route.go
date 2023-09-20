package routes

import (
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// RoutesWatch waits for routing update events and then probes the
// trusted https servers
type RoutesWatch struct {
	probes chan struct{}
	done   chan struct{}
}

// sendProbe sends a probe request over the probe channel
func (r *RoutesWatch) sendProbe() {
	select {
	case r.probes <- struct{}{}:
	case <-r.done:
	}
}

// start starts the RoutesWatch
func (r *RoutesWatch) start() {
	// register for route update events
	events := make(chan netlink.RouteUpdate)
	if err := netlink.RouteSubscribe(events, r.done); err != nil {
		log.WithError(err).Fatal("TND route subscribe error")
	}

	// run initial probe
	r.sendProbe()

	// handle route update events
	for e := range events {
		switch e.Type {
		case unix.RTM_NEWROUTE:
			log.WithField("dst", e.Dst).Debug("TND got route NEW event")
		case unix.RTM_DELROUTE:
			log.WithField("dst", e.Dst).Debug("TND got route DEL event")
		}
		r.sendProbe()
	}
}

// Start starts the RoutesWatch
func (r *RoutesWatch) Start() {
	go r.start()
}

// Stop stops the RoutesWatch
func (r *RoutesWatch) Stop() {
	// NOTE: this will not terminate the RoutesWatch goroutine until the
	// next netlink event arrives due to a known issue of the netlink
	// library we use. So, we cannot really wait for the goroutine
	// termination here
	close(r.done)
}

// Probes returns the probe channel
func (r *RoutesWatch) Probes() chan struct{} {
	return r.probes
}

// NewRoutesWatch returns a new RoutesWatch
func NewRoutesWatch(probes chan struct{}) *RoutesWatch {
	return &RoutesWatch{
		probes: probes,
		done:   make(chan struct{}),
	}
}
