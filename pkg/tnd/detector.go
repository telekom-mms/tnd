package tnd

import (
	"math/rand/v2"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/telekom-mms/tnd/internal/files"
	"github.com/telekom-mms/tnd/internal/https"
	"github.com/telekom-mms/tnd/internal/routes"
)

// Detector realizes the trusted network detection.
type Detector struct {
	config  *Config
	probes  chan struct{}
	results chan bool
	done    chan struct{}
	servers []*https.Server
	dialer  *net.Dialer

	// route and file watch
	rw routes.Watcher
	fw files.Watcher

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

// SetServers sets the https server urls and their expected hashes in the
// servers map as trusted servers; map key is the server url, value is the
// server's hash. Note: servers must be set before Start().
func (d *Detector) SetServers(servers map[string]string) {
	d.servers = []*https.Server{}
	for url, hash := range servers {
		server := https.NewServer(url, hash)
		d.servers = append(d.servers, server)
	}
}

// GetServers returns the https servers as map; map key is the server url,
// value is the server's hash.
func (d *Detector) GetServers() map[string]string {
	servers := make(map[string]string)
	for _, s := range d.servers {
		servers[s.URL] = s.Hash
	}
	return servers
}

// SetDialer sets a custom dialer for the https connections; note: the dialer
// must be set before Start().
func (d *Detector) SetDialer(dialer *net.Dialer) {
	d.dialer = dialer
}

// GetDialer returns the custom dialer for the https connections.
func (d *Detector) GetDialer() *net.Dialer {
	return d.dialer
}

// sendResult sends result over channel c.
func (d *Detector) sendResult(c chan bool, result bool) {
	select {
	case c <- result:
	case <-d.done:
	}
}

// probe checks the servers and sends the result back over probeResults.
func (d *Detector) probe() {
	for _, i := range rand.Perm(len(d.servers)) {
		s := d.servers[i]
		// sleep between server probes to let network settle a bit in
		// case of a burst of routing and dns changes, e.g, when
		// connecting to a new network
		time.Sleep(d.config.WaitCheck)

		if s.Check(d.dialer, d.config.HTTPSTimeout) {
			// TODO: be more strict and require all trusted servers
			// to be reachable?
			log.WithField("url", s.URL).Debug("TND https server trusted")
			d.sendResult(d.probeResults, true)
			return
		}
		log.WithField("url", s.URL).Debug("TND https server not trusted")
	}
	d.sendResult(d.probeResults, false)
}

// resetTimer resets the periodic probe timer.
func (d *Detector) resetTimer() {
	if d.trusted {
		d.timer.Reset(d.config.TrustedTimer)
	} else {
		d.timer.Reset(d.config.UntrustedTimer)
	}
}

// handleProbeRequest handles a probe request.
func (d *Detector) handleProbeRequest() {
	if d.running {
		d.runAgain = true
		return
	}
	d.running = true
	go d.probe()

}

// handleProbeResult handles the probe result r.
func (d *Detector) handleProbeResult(r bool) {
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
		return
	}
	if !d.timer.Stop() {
		<-d.timer.C
	}
	d.resetTimer()
}

// handleTimer handles a timer event.
func (d *Detector) handleTimer() {
	if !d.running && !d.runAgain {
		// no probes active, trigger new probe
		log.Debug("TND periodic probe timer")
		d.running = true
		go d.probe()
	}

	// reset timer
	d.resetTimer()
}

// start starts the trusted network detection.
func (d *Detector) start() {
	// signal stop to user via results
	defer close(d.results)
	defer d.rw.Stop()
	defer d.fw.Stop()

	// set timer for periodic checks
	d.timer = time.NewTimer(d.config.UntrustedTimer)

	// main loop
	for {
		select {
		case <-d.probes:
			d.handleProbeRequest()

		case r := <-d.probeResults:
			d.handleProbeResult(r)

		case <-d.timer.C:
			d.handleTimer()

		case <-d.done:
			if !d.timer.Stop() {
				<-d.timer.C
			}
			return
		}
	}
}

// Start starts the trusted network detection.
func (d *Detector) Start() error {
	// start route watching
	if err := d.rw.Start(); err != nil {
		return err
	}

	// start file watching
	if err := d.fw.Start(); err != nil {
		d.rw.Stop()
		return err
	}

	// start detector
	go d.start()
	return nil
}

// Stop stops the running TND.
func (d *Detector) Stop() {
	close(d.done)
	for range d.results {
		// wait for exit
		log.Debug("TND dropping result during shutdown")
	}
}

// Probe triggers a trusted network probe.
func (d *Detector) Probe() {
	select {
	case d.probes <- struct{}{}:
	case <-d.done:
	}
}

// Results returns the results channel.
func (d *Detector) Results() chan bool {
	return d.results
}

// NewDetector returns a new Detector.
func NewDetector(config *Config) *Detector {
	probes := make(chan struct{})
	return &Detector{
		config:  config,
		probes:  probes,
		results: make(chan bool),
		done:    make(chan struct{}),
		dialer:  &net.Dialer{},
		rw:      routes.NewWatch(probes),
		fw:      files.NewWatch(probes),

		probeResults: make(chan bool),
	}
}
