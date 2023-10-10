package routes

import "testing"

// TestWatchStartStop tests Start and Stop of Watch.
func TestWatchStartStop(_ *testing.T) {
	probes := make(chan struct{})
	rw := NewWatch(probes)
	rw.Start()
	rw.Stop()
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
	if rw.probes != probes {
		t.Errorf("got %p, want %p", rw.probes, probes)
	}
	if rw.done == nil {
		t.Errorf("got nil, want != nil")
	}
}
