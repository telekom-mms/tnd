// Package tndtest contains a trusted network detector for TND testing.
package tndtest

import (
	"net"
)

// Funcs are functions used by Detector for use in tests.
type Funcs struct {
	SetServers func(servers map[string]string)
	GetServers func() map[string]string
	SetDialer  func(dialer *net.Dialer)
	GetDialer  func() *net.Dialer
	Start      func() error
	Stop       func()
	Probe      func()
	Results    func() chan bool
}

// Detector is a simple Detector for use in tests.
type Detector struct {
	Funcs Funcs
}

// SetServers sets the https server urls and expected hashes as the list of
// trusted servers. Map key is the server url, value is the hash.
func (d *Detector) SetServers(servers map[string]string) {
	if d.Funcs.SetServers != nil {
		d.Funcs.SetServers(servers)
	}
}

// GetServers returns the trusted server urls and hashes.
func (d *Detector) GetServers() map[string]string {
	if d.Funcs.GetServers != nil {
		return d.Funcs.GetServers()
	}
	return nil
}

// SetDialer sets a custom dialer for the https connections.
func (d *Detector) SetDialer(dialer *net.Dialer) {
	if d.Funcs.SetDialer != nil {
		d.Funcs.SetDialer(dialer)
	}
}

// GetDialer returns the custom dialer for the https connections.
func (d *Detector) GetDialer() *net.Dialer {
	if d.Funcs.GetDialer != nil {
		return d.Funcs.GetDialer()
	}
	return nil
}

// Start starts the trusted network detection.
func (d *Detector) Start() error {
	if d.Funcs.Start != nil {
		return d.Funcs.Start()
	}
	return nil
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
