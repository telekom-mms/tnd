package tnd

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"reflect"
	"testing"
)

// testWatcher is a watcher that implements the routes.Watcher and
// files.Watcher interfaces.
type testWatcher struct{ err error }

func (t *testWatcher) Start() error          { return t.err }
func (t *testWatcher) Stop()                 {}
func (t *testWatcher) Probes() chan struct{} { return nil }

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
func TestTNDStartStop(t *testing.T) {
	// test rw error
	t.Run("routes watch error", func(t *testing.T) {
		tnd := NewDetector(NewConfig())
		tnd.rw = &testWatcher{err: errors.New("test error")}
		tnd.fw = &testWatcher{}
		if err := tnd.Start(); err == nil {
			t.Error("start should fail")
			return
		}
	})

	// test fw error
	t.Run("files watch error", func(t *testing.T) {
		tnd := NewDetector(NewConfig())
		tnd.rw = &testWatcher{}
		tnd.fw = &testWatcher{err: errors.New("test error")}
		if err := tnd.Start(); err == nil {
			t.Error("start should fail")
			return
		}
	})

	// test without errors
	t.Run("no errors", func(t *testing.T) {
		tnd := NewDetector(NewConfig())
		tnd.rw = &testWatcher{}
		tnd.fw = &testWatcher{}
		if err := tnd.Start(); err != nil {
			t.Errorf("start should not fail: %v", err)
			return
		}
		tnd.Stop()
	})
}

// TestTNDProbe tests Probe of TND.
func TestTNDProbe(t *testing.T) {
	tnd := NewDetector(NewConfig())
	tnd.rw = &testWatcher{}
	tnd.fw = &testWatcher{}
	if err := tnd.Start(); err != nil {
		t.Fatal(err)
	}
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
	c := NewConfig()
	tnd := NewDetector(c)

	if tnd.config != c {
		t.Errorf("got %v, want %v", tnd.config, c)
	}

	for i, x := range []any{
		tnd.probes,
		tnd.results,
		tnd.done,
		tnd.dialer,
		tnd.rw,
		tnd.fw,
		tnd.probeResults,
	} {
		if x == nil {
			t.Errorf("got nil, want != nil: %d", i)
		}
	}
}
