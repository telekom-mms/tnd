// Package tnd contains components for trusted network detection.
package tnd

import (
	"net"
)

// TND is the trusted network detection.
type TND interface {
	SetServers(map[string]string)
	GetServers() map[string]string
	SetDialer(dialer *net.Dialer)
	GetDialer() *net.Dialer
	Start()
	Stop()
	Probe()
	Results() chan bool
}
