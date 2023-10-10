package files

import "testing"

// TestWatchStartStop tests Start() and Stop() of Watch.
func TestWatchStartStop(_ *testing.T) {
	probes := make(chan struct{})
	fw := NewWatch(probes)
	fw.Start()
	fw.Stop()
}

// TestNewWatch tests NewWatch.
func TestNewWatch(t *testing.T) {
	probes := make(chan struct{})
	fw := NewWatch(probes)
	if fw.probes != probes {
		t.Errorf("got %p, want %p", fw.probes, probes)
	}
	if fw.done == nil {
		t.Errorf("got nil, want != nil")
	}
	if fw.closed == nil {
		t.Errorf("got nil, want != nil")
	}
}
