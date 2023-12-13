package tndtest

import (
	"net"
	"reflect"
	"testing"
)

// TestDetectorSetGetServers tests SetServers and GetServers of Detector.
func TestDetectorSetGetServers(t *testing.T) {
	d := NewDetector()

	// test no func set
	url := "https://example.com"
	hash := "abcdefabcdefabcdefabcdef"
	servers := map[string]string{url: hash}
	d.SetServers(servers)
	if d.GetServers() != nil {
		t.Errorf("servers should be nil")
	}

	// test func set
	testServers := map[string]string{}
	d.Funcs.SetServers = func(s map[string]string) {
		testServers = s
	}
	d.Funcs.GetServers = func() map[string]string {
		return testServers
	}
	d.SetServers(servers)
	got := d.GetServers()
	if !reflect.DeepEqual(got, servers) {
		t.Errorf("got %v, want %v", got, servers)
	}
}

// TestDetectorSetDialer tests SetDialer and GetDialer of Detector.
func TestDetectorSetGetDialer(t *testing.T) {
	d := NewDetector()

	// test no func set
	d.SetDialer(&net.Dialer{})
	if d.GetDialer() != nil {
		t.Errorf("dialer should be nil")
	}

	// test func set
	want := &net.Dialer{}
	testDialer := &net.Dialer{}
	d.Funcs.SetDialer = func(dialer *net.Dialer) {
		testDialer = dialer
	}
	d.Funcs.GetDialer = func() *net.Dialer {
		return testDialer
	}

	d.SetDialer(want)
	got := d.GetDialer()
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
	d.Funcs.Start = func() error {
		got = true
		return nil
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
