package trustnet

import (
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/telekom-mms/tnd/internal/trusthttps"
)

// TND realizes the trusted network detection
type TND struct {
	https   *trusthttps.TrustedHTTPS
	results chan bool
	done    chan struct{}
}

// AddServer adds a trusted http server url and expected hash
func (t *TND) AddServer(url, hash string) {
	t.https.AddServer(url, hash)
}

// SetDialer sets a custom dialer for http connections
func (t *TND) SetDialer(dialer *net.Dialer) {
	t.https.SetDialer(dialer)
}

// sendResult sends result over the result channel
func (t *TND) sendResult(result bool) {
	select {
	case t.results <- result:
	case <-t.done:
	}
}

// start starts the trusted network detection
func (t *TND) start() {
	// signal stop to user via results
	defer close(t.results)

	// start trusted https
	t.https.Start()
	defer t.https.Stop()

	// start route watching
	rw := trusthttps.NewRoutesWatch(t.https.Probes())
	rw.Start()
	defer rw.Stop()

	// start file watching
	fw := trusthttps.NewFilesWatch(t.https.Probes())
	fw.Start()
	defer fw.Stop()

	// run main loop
	for {
		select {
		case r, ok := <-t.https.Results():
			if !ok {
				return
			}
			log.WithField("trusted", r).Debug("TND https result")
			t.sendResult(r)

		case <-t.done:
			return
		}
	}
}

// Start starts the trusted network detection
func (t *TND) Start() {
	go t.start()
}

// Stop stops the running TND
func (t *TND) Stop() {
	close(t.done)
	for range t.results {
		// wait for exit
	}
}

// Results returns the results channel
func (t *TND) Results() chan bool {
	return t.results
}

// NewTND returns a new TND
func NewTND() *TND {
	return &TND{
		https:   trusthttps.NewTrustedHTTPS(),
		results: make(chan bool),
		done:    make(chan struct{}),
	}
}
