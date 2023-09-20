package trustnet

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/telekom-mms/tnd/internal/files"
	"github.com/telekom-mms/tnd/internal/https"
	"github.com/telekom-mms/tnd/internal/routes"
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

// TND realizes the trusted network detection
type TND struct {
	probes  chan struct{}
	results chan bool
	done    chan struct{}
	servers []*https.Server
	dialer  *net.Dialer

	// timer
	timer *time.Timer

	// probe result channel and probe function
	probeResults chan bool

	// is the network trusted, are probes currently running or
	// have to run again?
	trusted  bool
	running  bool
	runAgain bool
}

// AddServer adds the https server url and its expected hash to the list of
// trusted servers; note: all servers must be added before Start()
func (t *TND) AddServer(url, hash string) {
	server := https.NewServer(url, hash)
	t.servers = append(t.servers, server)
}

// SetDialer sets a custom dialer for the https connections; note: the dialer
// must be set before Start()
func (t *TND) SetDialer(dialer *net.Dialer) {
	t.dialer = dialer
}

// sendResult sends result over channel c
func (t *TND) sendResult(c chan bool, result bool) {
	select {
	case c <- result:
	case <-t.done:
	}
}

// probe checks the servers and sends the result back over probeResults
func (t *TND) probe() {
	for _, s := range t.servers {
		// sleep a second between server probes to let network
		// settle a bit in case of a burst of routing and dns
		// changes, e.g, when connecting to a new network
		time.Sleep(waitCheck * time.Second)

		if s.Check(t.dialer, httpsTimeout*time.Second) {
			// TODO: be more strict and require all trusted servers
			// to be reachable?
			// TODO: probe servers in random order?
			log.WithField("url", s.URL).Debug("TND https server trusted")
			t.sendResult(t.probeResults, true)
			return
		}
		log.WithField("url", s.URL).Debug("TND https server not trusted")
	}
	t.sendResult(t.probeResults, false)
}

// resetTimer resets the periodic probe timer
func (t *TND) resetTimer() {
	if t.trusted {
		t.timer.Reset(trustedTimer * time.Second)
	} else {
		t.timer.Reset(untrustedTimer * time.Second)
	}
}

// start starts the trusted network detection
func (t *TND) start() {
	// signal stop to user via results
	defer close(t.results)

	// start route watching
	rw := routes.NewRoutesWatch(t.probes)
	rw.Start()
	defer rw.Stop()

	// start file watching
	fw := files.NewFilesWatch(t.probes)
	fw.Start()
	defer fw.Stop()

	// set timer for periodic checks
	t.timer = time.NewTimer(untrustedTimer * time.Second)

	// main loop
	for {
		select {
		case <-t.probes:
			if t.running {
				t.runAgain = true
				break
			}
			t.running = true
			go t.probe()

		case r := <-t.probeResults:
			// handle probe result
			t.running = false
			if t.runAgain {
				// we must trigger another probe
				t.runAgain = false
				t.running = true
				go t.probe()
			}
			log.WithField("trusted", r).Debug("TND https result")
			t.trusted = r
			t.sendResult(t.results, r)

			// reset periodic probing timer
			if t.running {
				// probing still active and new results about
				// to arrive, so wait for them before resetting
				// the timer
				break
			}
			if !t.timer.Stop() {
				<-t.timer.C
			}
			t.resetTimer()

		case <-t.timer.C:
			if !t.running && !t.runAgain {
				// no probes active, trigger new probe
				log.Debug("TND periodic probe timer")
				t.running = true
				go t.probe()
			}

			// reset timer
			t.resetTimer()

		case <-t.done:
			if !t.timer.Stop() {
				<-t.timer.C
			}
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
		probes:  make(chan struct{}),
		results: make(chan bool),
		done:    make(chan struct{}),
		dialer:  &net.Dialer{},

		probeResults: make(chan bool),
	}
}
