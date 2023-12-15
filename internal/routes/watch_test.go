package routes

import (
	"errors"
	"testing"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// TestWatchStartEvents tests start of Watch, events.
func TestWatchStartEvents(_ *testing.T) {
	// create and start watch
	probes := make(chan struct{})
	rw := NewWatch(probes)
	go rw.start()
	<-probes

	// new route event
	rw.events <- netlink.RouteUpdate{Type: unix.RTM_NEWROUTE}
	<-probes

	// delete route event
	rw.events <- netlink.RouteUpdate{Type: unix.RTM_DELROUTE}
	<-probes

	close(rw.done)
}

// TestWatchStartStop tests Start and Stop of Watch.
func TestWatchStartStop(t *testing.T) {
	probes := make(chan struct{})

	t.Run("subscribe error", func(t *testing.T) {
		defer func() { netlinkRouteSubscribe = netlink.RouteSubscribe }()
		netlinkRouteSubscribe = func(chan<- netlink.RouteUpdate, <-chan struct{}) error {
			return errors.New("test error")
		}

		rw := NewWatch(probes)
		if err := rw.Start(); err == nil {
			t.Error("start should fail")
		}
	})

	t.Run("no errors", func(t *testing.T) {
		rw := NewWatch(probes)
		if err := rw.Start(); err != nil {
			t.Errorf("start should not fail: %v", err)
		}
		rw.Stop()
	})
}

// TestWatchProbes tests Probes of Watch.
func TestWatchProbes(t *testing.T) {
	probes := make(chan struct{})
	rw := NewWatch(probes)
	if rw.Probes() != probes {
		t.Errorf("got %p, want %p", rw.Probes(), probes)
	}
}

// TestNewWatch tests NewWatch.
func TestNewWatch(t *testing.T) {
	probes := make(chan struct{})
	rw := NewWatch(probes)
	if rw.events == nil {
		t.Errorf("got nil, want != nil")
	}
	if rw.probes != probes {
		t.Errorf("got %p, want %p", rw.probes, probes)
	}
	if rw.done == nil {
		t.Errorf("got nil, want != nil")
	}
}
