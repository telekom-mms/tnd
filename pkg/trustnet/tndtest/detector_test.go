package tndtest

import (
	"net"
	"testing"
)

// TestDetectorAddServer tests AddServer of Detector.
func TestDetectorAddServer(t *testing.T) {
	d := NewDetector()

	// test no func set
	url := "https://example.com"
	hash := "abcdefabcdefabcdefabcdef"
	d.AddServer(url, hash)

	// test func set
	gotURL := ""
	gotHash := ""
	d.Funcs.AddServer = func(url, hash string) {
		gotURL = url
		gotHash = hash
	}
	d.AddServer(url, hash)
	if gotURL != url || gotHash != hash {
		t.Errorf("got %s %s, want %s %s", gotURL, gotHash, url, hash)
	}
}

// TestDetectorSetDialer tests SetDialer of Detector.
func TestDetectorSetDialer(t *testing.T) {
	d := NewDetector()

	// test no func set
	d.SetDialer(&net.Dialer{})

	// test func set
	want := &net.Dialer{}
	got := &net.Dialer{}
	d.Funcs.SetDialer = func(dialer *net.Dialer) {
		got = dialer
	}
	d.SetDialer(want)
	if got != want {
		t.Errorf("got %p, want %p", got, want)
	}
}

// TestDetectorStart tests Start of Detector.
func TestDetectorStart(t *testing.T) {
	d := NewDetector()

	// test no func set
	d.Start()

	// test func set
	want := true
	got := false
	d.Funcs.Start = func() {
		got = true
	}
	d.Start()
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

// TestDetectorStop tests Stop of Detector.
func TestDetectorStop(t *testing.T) {
	d := NewDetector()

	// test no func set
	d.Stop()

	// test func set
	want := true
	got := false
	d.Funcs.Stop = func() {
		got = true
	}
	d.Stop()
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

// TestDetectorProbe tests Probe of Detector.
func TestDetectorProbe(t *testing.T) {
	d := NewDetector()

	// test no func set
	d.Probe()

	// test func set
	want := true
	got := false
	d.Funcs.Probe = func() {
		got = true
	}
	d.Probe()
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

// TestDetectorResults tests Results of Detector.
func TestDetectorResults(t *testing.T) {
	d := NewDetector()

	// test no func set
	if d.Results() != nil {
		t.Errorf("got unexpected results channel")
	}

	// test func set
	want := make(chan bool)
	got := make(chan bool)
	d.Funcs.Results = func() chan bool {
		got = want
		return nil
	}
	d.Results()
	if got != want {
		t.Errorf("got %p, want %p", got, want)
	}
}

// TestNewDetector tests NewDetector.
func TestNewDetector(t *testing.T) {
	if NewDetector() == nil {
		t.Errorf("invalid detector")
	}
}
