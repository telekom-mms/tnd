package trustnet

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/telekom-mms/tnd/internal/files"
	"github.com/telekom-mms/tnd/internal/https"
	"github.com/telekom-mms/tnd/internal/routes"
)

// Detector realizes the trusted network detection
type Detector struct {
	config  *Config
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
func (d *Detector) AddServer(url, hash string) {
	server := https.NewServer(url, hash)
	d.servers = append(d.servers, server)
}

// SetDialer sets a custom dialer for the https connections; note: the dialer
// must be set before Start()
func (d *Detector) SetDialer(dialer *net.Dialer) {
	d.dialer = dialer
}

// sendResult sends result over channel c
func (d *Detector) sendResult(c chan bool, result bool) {
	select {
	case c <- result:
	case <-d.done:
	}
}

// probe checks the servers and sends the result back over probeResults
func (d *Detector) probe() {
	for _, s := range d.servers {
		// sleep a second between server probes to let network
		// settle a bit in case of a burst of routing and dns
		// changes, e.g, when connecting to a new network
		time.Sleep(d.config.WaitCheck)

		if s.Check(d.dialer, d.config.HTTPSTimeout) {
			// TODO: be more strict and require all trusted servers
			// to be reachable?
			// TODO: probe servers in random order?
			log.WithField("url", s.URL).Debug("TND https server trusted")
			d.sendResult(d.probeResults, true)
			return
		}
		log.WithField("url", s.URL).Debug("TND https server not trusted")
	}
	d.sendResult(d.probeResults, false)
}

// resetTimer resets the periodic probe timer
func (d *Detector) resetTimer() {
	if d.trusted {
		d.timer.Reset(d.config.TrustedTimer)
	} else {
		d.timer.Reset(d.config.UntrustedTimer)
	}
}

// start starts the trusted network detection
func (d *Detector) start() {
	// signal stop to user via results
	defer close(d.results)

	// start route watching
	rw := routes.NewRoutesWatch(d.probes)
	rw.Start()
	defer rw.Stop()

	// start file watching
	fw := files.NewFilesWatch(d.probes)
	fw.Start()
	defer fw.Stop()

	// set timer for periodic checks
	d.timer = time.NewTimer(d.config.UntrustedTimer)

	// main loop
	for {
		select {
		case <-d.probes:
			if d.running {
				d.runAgain = true
				break
			}
			d.running = true
			go d.probe()

		case r := <-d.probeResults:
			// handle probe result
			d.running = false
			if d.runAgain {
				// we must trigger another probe
				d.runAgain = false
				d.running = true
				go d.probe()
			}
			log.WithField("trusted", r).Debug("TND https result")
			d.trusted = r
			d.sendResult(d.results, r)

			// reset periodic probing timer
			if d.running {
				// probing still active and new results about
				// to arrive, so wait for them before resetting
				// the timer
				break
			}
			if !d.timer.Stop() {
				<-d.timer.C
			}
			d.resetTimer()

		case <-d.timer.C:
			if !d.running && !d.runAgain {
				// no probes active, trigger new probe
				log.Debug("TND periodic probe timer")
				d.running = true
				go d.probe()
			}

			// reset timer
			d.resetTimer()

		case <-d.done:
			if !d.timer.Stop() {
				<-d.timer.C
			}
			return
		}
	}
}

// Start starts the trusted network detection
func (d *Detector) Start() {
	go d.start()
}

// Stop stops the running TND
func (d *Detector) Stop() {
	close(d.done)
	for range d.results {
		// wait for exit
		log.Debug("TND dropping result during shutdown")
	}
}

// Probe triggers a trusted network probe
func (d *Detector) Probe() {
	select {
	case d.probes <- struct{}{}:
	case <-d.done:
	}
}

// Results returns the results channel
func (d *Detector) Results() chan bool {
	return d.results
}

// NewDetector returns a new Detector
func NewDetector(config *Config) *Detector {
	return &Detector{
		config:  config,
		probes:  make(chan struct{}),
		results: make(chan bool),
		done:    make(chan struct{}),
		dialer:  &net.Dialer{},

		probeResults: make(chan bool),
	}
}
