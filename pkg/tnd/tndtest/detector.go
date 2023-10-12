// Package tndtest contains a trusted network detector for TND testing.
package tndtest

import (
	"net"
)

// Funcs are functions used by Detector for use in tests.
type Funcs struct {
	AddServer func(url, hash string)
	SetDialer func(dialer *net.Dialer)
	Start     func()
	Stop      func()
	Probe     func()
	Results   func() chan bool
}

// Detector is a simple Detector for use in tests.
type Detector struct {
	Funcs Funcs
}

// AddServer adds the https server url and its expected hash to the list of
// trusted servers.
func (d *Detector) AddServer(url, hash string) {
	if d.Funcs.AddServer != nil {
		d.Funcs.AddServer(url, hash)
	}
}

// SetDialer sets a custom dialer for the https connections.
func (d *Detector) SetDialer(dialer *net.Dialer) {
	if d.Funcs.SetDialer != nil {
		d.Funcs.SetDialer(dialer)
	}
}

// Start starts the trusted network detection.
func (d *Detector) Start() {
	if d.Funcs.Start != nil {
		d.Funcs.Start()
	}
}

// Stop stops the running TND.
func (d *Detector) Stop() {
	if d.Funcs.Stop != nil {
		d.Funcs.Stop()
	}
}

// Probe triggers a trusted network probe.
func (d *Detector) Probe() {
	if d.Funcs.Probe != nil {
		d.Funcs.Probe()
	}
}

// Results returns the results channel.
func (d *Detector) Results() chan bool {
	if d.Funcs.Results != nil {
		return d.Funcs.Results()
	}
	return nil
}

// NewDetector returns a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}
