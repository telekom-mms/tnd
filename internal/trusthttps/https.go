package trusthttps

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

const (
	// waitCheck is the wait time before http checks in seconds
	waitCheck = 1

	// httpsTimeout is the timeout for http requests in seconds
	httpsTimeout = 5

	// untrustedTimer is the timer for periodic checks in case of an
	// untrusted network in seconds
	untrustedTimer = 30

	// trustedTimer is the timer for periodic checks in case of a
	// trusted network in seconds
	trustedTimer = 60
)

// trustedHTTPSServer is a trusted https server and its certificate hash
type trustedHTTPSServer struct {
	url  string
	hash string
}

// check probes the https server and checks the certificate hash using dialer
func (t *trustedHTTPSServer) check(dialer *net.Dialer) bool {
	// connect to server
	tr := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   httpsTimeout * time.Second,
	}
	r, err := client.Head(t.url)
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
	if fp != t.hash {
		log.WithFields(log.Fields{
			"got":  fp,
			"want": t.hash,
		}).Debug("TND https server hash mismatch")
		return false
	}

	// all checks passed
	return true
}

// TrustedHTTPS stores and checks trusted https servers
type TrustedHTTPS struct {
	probes  chan struct{}
	results chan bool
	done    chan struct{}
	servers []*trustedHTTPSServer
	dialer  *net.Dialer
}

// AddServer adds the https server url and its expected hash to the list of
// trusted servers; note: all servers must be added before Start()
func (t *TrustedHTTPS) AddServer(url, hash string) {
	server := &trustedHTTPSServer{
		url,
		strings.ToLower(hash),
	}
	t.servers = append(t.servers, server)
}

// SetDialer sets a custom dialer for the https connections; note: the dialer
// must be set before Start()
func (t *TrustedHTTPS) SetDialer(dialer *net.Dialer) {
	t.dialer = dialer
}

// Probe checks if trusted https servers are reachable and their certificate
// hashes match
func (t *TrustedHTTPS) Probe() {
	t.probes <- struct{}{}
}

// start starts running the TrustedHTTPS detection
func (t *TrustedHTTPS) start() {
	defer close(t.results)

	// helper for sending results
	sendResult := func(c chan bool, result bool) {
		// send result over channel c or abort
		select {
		case c <- result:
		case <-t.done:
		}
	}

	// probe result channel and probe function
	probeResults := make(chan bool)
	probeFunc := func() {
		for _, s := range t.servers {
			// sleep a second between server probes to let network
			// settle a bit in case of a burst of routing and dns
			// changes, e.g, when connecting to a new network
			time.Sleep(waitCheck * time.Second)

			if s.check(t.dialer) {
				// TODO: be more strict and require all trusted servers
				// to be reachable?
				// TODO: probe servers in random order?
				log.WithField("url", s.url).Debug("TND https server trusted")
				sendResult(probeResults, true)
				return
			}
			log.WithField("url", s.url).Debug("TND https server not trusted")
		}
		sendResult(probeResults, false)
	}

	// is the network trusted, are probes currently running or
	// have to run again?
	trusted := false
	running := false
	runAgain := false

	// set timer for periodic checks
	timer := time.NewTimer(untrustedTimer * time.Second)
	resetTimer := func() {
		if trusted {
			timer.Reset(trustedTimer * time.Second)
		} else {
			timer.Reset(untrustedTimer * time.Second)
		}
	}

	for {
		select {
		case <-t.probes:
			if running {
				runAgain = true
				break
			}
			running = true
			go probeFunc()

		case r := <-probeResults:
			// handle probe result
			running = false
			if runAgain {
				// we must trigger another probe
				runAgain = false
				running = true
				go probeFunc()
			}
			trusted = r
			sendResult(t.results, r)

			// reset periodic probing timer
			if running {
				// probing still active and new results about
				// to arrive, so wait for them before resetting
				// the timer
				break
			}
			if !timer.Stop() {
				<-timer.C
			}
			resetTimer()

		case <-timer.C:
			if !running && !runAgain {
				// no probes active, trigger new probe
				log.Debug("TND periodic probe timer")
				running = true
				go probeFunc()
			}

			// reset timer
			resetTimer()

		case <-t.done:
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}
}

// Start starts running the TrustedHTTPS detection
func (t *TrustedHTTPS) Start() {
	go t.start()
}

// Stop stops running the trusted https detection
func (t *TrustedHTTPS) Stop() {
	close(t.done)
	for range t.results {
		// wait for channel close
	}
}

// Probes returns the probes channel
func (t *TrustedHTTPS) Probes() chan struct{} {
	return t.probes
}

// Results returns the results channel
func (t *TrustedHTTPS) Results() chan bool {
	return t.results
}

// NewTrustedHTTPS returns a new TrustedHTTPS
func NewTrustedHTTPS() *TrustedHTTPS {
	return &TrustedHTTPS{
		probes:  make(chan struct{}),
		results: make(chan bool),
		done:    make(chan struct{}),
		dialer:  &net.Dialer{},
	}
}
