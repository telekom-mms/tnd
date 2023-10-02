package tnd

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"testing"
)

// TestTNDAddServer tests AddServer of TND.
func TestTNDAddServer(t *testing.T) {
	tnd := NewDetector(NewConfig())
	url := "http://test.example.com:442"
	cert := []byte("raw test certificate")
	hash := sha256.Sum256(cert)
	hs := hex.EncodeToString(hash[:])
	tnd.AddServer(url, hs)

	want := url
	got := tnd.servers[0].URL
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = hs
	got = tnd.servers[0].Hash
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestTNDSetDialer tests SetDialer of TND.
func TestTNDSetDialer(t *testing.T) {
	tnd := NewDetector(NewConfig())
	dialer := &net.Dialer{}
	tnd.SetDialer(dialer)

	want := dialer
	got := tnd.dialer
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
