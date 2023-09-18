package https

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Server is a trusted https server and its certificate hash
type Server struct {
	URL  string
	Hash string
}

// Check probes the https server and checks the certificate hash using dialer
func (s *Server) Check(dialer *net.Dialer, timeout time.Duration) bool {
	// connect to server
	tr := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	r, err := client.Head(s.URL)
	if err != nil {
		log.WithError(err).Debug("TND http HEAD request error")
		return false
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.WithError(err).Error("TND could not close http response body")
		}
	}()
	if _, err := io.Copy(ioutil.Discard, r.Body); err != nil {
		log.WithError(err).Error("TND could not read http response body")
	}

	// make sure we created an tls connection
	if r.TLS == nil {
		log.WithField("error", "no tls connection to https server").
			Debug("TND http connection error")
		return false
	}

	// get certificate and the fingerprint
	cert := r.TLS.PeerCertificates[0]
	hash := sha256.Sum256(cert.Raw)
	fp := hex.EncodeToString(hash[:])

	// check if fingerprint matches expected hash
	if fp != s.Hash {
		log.WithFields(log.Fields{
			"got":  fp,
			"want": s.Hash,
		}).Debug("TND https server hash mismatch")
		return false
	}

	// all checks passed
	return true
}

// NewServer returns a new Server with url and hash
func NewServer(url, hash string) *Server {
	return &Server{
		URL:  url,
		Hash: strings.ToLower(hash),
	}
}
