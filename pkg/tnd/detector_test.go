package tnd

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"reflect"
	"testing"
)

// TestDetectorSetGetServers tests SetServers and GetServers of Detector.
func TestDetectorSetGetServers(t *testing.T) {
	tnd := NewDetector(NewConfig())

	url := "http://test.example.com:442"
	cert := []byte("raw test certificate")
	hash := sha256.Sum256(cert)
	hs := hex.EncodeToString(hash[:])
	want := map[string]string{url: hs}

	tnd.SetServers(want)
	got := tnd.GetServers()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestDetectorSetGetDialer tests SetDialer and GetDialer of Detector.
func TestDetectorSetGetDialer(t *testing.T) {
	tnd := NewDetector(NewConfig())
	dialer := &net.Dialer{}
	tnd.SetDialer(dialer)

	want := dialer
	got := tnd.GetDialer()
	if got != want {
		t.Errorf("got %p, want %p", got, want)
	}
}

// TestTNDStartStop tests Start and Stop of TND.
func TestTNDStartStop(_ *testing.T) {
	tnd := NewDetector(NewConfig())
	tnd.Start()
	tnd.Stop()
}

// TestTNDProbe tests Probe of TND.
func TestTNDProbe(t *testing.T) {
	tnd := NewDetector(NewConfig())
	tnd.Start()
	tnd.Probe()
	want := false
	got := <-tnd.Results()
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
	tnd.Stop()
}

// TestTNDResults tests Results of TND.
func TestTNDResults(t *testing.T) {
	tnd := NewDetector(NewConfig())
	want := tnd.results
	got := tnd.Results()
	if want != got {
		t.Errorf("got %p, want %p", got, want)
	}
}

// TestNewTND tests NewTND.
func TestNewTND(t *testing.T) {
	tnd := NewDetector(NewConfig())
	if tnd.probes == nil {
		t.Errorf("got nil, want != nil")
	}
	if tnd.results == nil {
		t.Errorf("got nil, want != nil")
	}
	if tnd.done == nil {
		t.Errorf("got nil, want != nil")
	}
	if tnd.dialer == nil {
		t.Errorf("got nil, want != nil")
	}
}
