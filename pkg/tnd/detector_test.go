package tnd

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
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

// TestDetectorProbeHelper tests probe of Detector.
func TestDetectorProbeHelper(t *testing.T) {
	// start test https server
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {}))
	defer ts.Close()

	// create detector
	tnd := NewDetector(NewConfig())

	// test untrusted
	tnd.SetServers(map[string]string{ts.URL: "invalid"})
	go tnd.probe()

	want := false
	got := <-tnd.probeResults
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}

	// test trusted
	cert := ts.Certificate()
	sha := sha256.Sum256(cert.Raw)
	hash := hex.EncodeToString(sha[:])
	tnd.SetServers(map[string]string{ts.URL: hash})
	go tnd.probe()

	want = true
	got = <-tnd.probeResults
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

// TestDetectorHandleProbeRequest tests handleProbeRequest of Detector.
func TestDetectorHandleProbeRequest(t *testing.T) {
	// create detector
	tnd := NewDetector(NewConfig())

	// already running
	tnd.running = true
	tnd.handleProbeRequest()
	if tnd.runAgain != true {
		t.Error("run again should be true")
	}

	// not runnnig
	tnd.running = false
	tnd.handleProbeRequest()
	if tnd.running != true {
		t.Error("running should be true")
	}

	close(tnd.done)
}

// TestDetectorHandleProbeResult tests handleProbeResult of Detector.
func TestDetectorHandleProbeResult(t *testing.T) {
	// create detector
	tnd := NewDetector(NewConfig())

	// expire timer
	tnd.timer = time.NewTimer(0)

	// drain results channel
	go func() {
		for r := range tnd.results {
			log.Println("result:", r)
		}
	}()

	// test not trusted
	tnd.running = true
	tnd.handleProbeResult(false)
	if tnd.running != false {
		t.Error("running should be false")
	}

	// test trusted
	tnd.running = true
	tnd.handleProbeResult(true)
	if tnd.running != false {
		t.Error("running should be false")
	}

	// test with runAgain
	tnd.running = true
	tnd.runAgain = true
	tnd.handleProbeResult(false)
	if tnd.runAgain != false {
		t.Error("runAgain should be false")
	}

	close(tnd.results)
}

// TestDetectorHandleTimer tests handleTimer of Detector.
func TestDetectorHandleTimer(t *testing.T) {
	// create detector
	tnd := NewDetector(NewConfig())

	// expire timer
	tnd.timer = time.NewTimer(0)

	// test without already running
	tnd.handleTimer()
	if tnd.running != true {
		t.Error("running should be true")
	}

	// test with already running
	tnd.handleTimer()
	if tnd.running != true {
		t.Error("running should be true")
	}
}

// TestDetectorStartStop tests Start and Stop of Detector.
func TestDetectorStartStop(t *testing.T) {
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

// TestDetectorProbe tests Probe of Detector.
func TestDetectorProbe(t *testing.T) {
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

// TestDetectorResults tests Results of Detector.
func TestDetectorResults(t *testing.T) {
	tnd := NewDetector(NewConfig())
	want := tnd.results
	got := tnd.Results()
	if want != got {
		t.Errorf("got %p, want %p", got, want)
	}
}

// TestNewDetector tests NewDetector.
func TestNewDetector(t *testing.T) {
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
