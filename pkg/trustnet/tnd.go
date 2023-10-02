// Package trustnet contains components for trusted network detection.
package trustnet

import (
	"net"
)

// TND is the trusted network detection
type TND interface {
	AddServer(url, hash string)
	SetDialer(dialer *net.Dialer)
	Start()
	Stop()
	Probe()
	Results() chan bool
}
