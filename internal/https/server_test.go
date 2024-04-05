package https

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestServerCheck tests Check of Server.
func TestServerCheck(t *testing.T) {
	// start test https server
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {}))
	defer ts.Close()

	// test invalid server
	s := &Server{}
	want := false
	got := s.Check(&net.Dialer{}, time.Second)
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}

	// test invalid hash
	s = &Server{
		URL:  ts.URL,
		Hash: "",
	}
	want = false
	got = s.Check(&net.Dialer{}, time.Second)
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}

	// test valid hash
	cert := ts.Certificate()
	sha := sha256.Sum256(cert.Raw)
	hash := hex.EncodeToString(sha[:])
	s = &Server{
		URL:  ts.URL,
		Hash: hash,
	}
	want = true
	got = s.Check(&net.Dialer{}, time.Second)
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

// TestNewServer tests NewServer.
func TestNewServer(t *testing.T) {
	url := "test.example.com"
	hash := "test-hash"
	s := NewServer(url, hash)
	if s == nil ||
		s.URL != url ||
		s.Hash != hash {
		t.Errorf("invalid server")
	}
}
