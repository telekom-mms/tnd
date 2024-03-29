package tnd

import "time"

var (
	// WaitCheck is the wait time before http checks.
	WaitCheck = 1 * time.Second

	// HTTPSTimeout is the timeout for http requests.
	HTTPSTimeout = 5 * time.Second

	// UntrustedTimer is the timer for periodic checks in case of an
	// untrusted network.
	UntrustedTimer = 30 * time.Second

	// TrustedTimer is the timer for periodic checks in case of a
	// trusted network.
	TrustedTimer = 60 * time.Second
)

// Config is a TND configuration.
type Config struct {
	WaitCheck      time.Duration
	HTTPSTimeout   time.Duration
	UntrustedTimer time.Duration
	TrustedTimer   time.Duration
}

// Valid returns whether Config is valid.
func (c *Config) Valid() bool {
	if c == nil ||
		c.WaitCheck < 0 ||
		c.HTTPSTimeout < 0 ||
		c.UntrustedTimer < 0 ||
		c.TrustedTimer < 0 {
		// invalid
		return false
	}
	return true
}

// NewConfig returns a new Config.
func NewConfig() *Config {
	return &Config{
		WaitCheck:      WaitCheck,
		HTTPSTimeout:   HTTPSTimeout,
		UntrustedTimer: UntrustedTimer,
		TrustedTimer:   TrustedTimer,
	}
}
