package tnd

import "time"

var (
	// WatchFiles are the default files to watch for changes. They are
	// resolv.conf files in /etc and /run/systemd/resolve.
	WatchFiles = []string{
		"/etc/resolv.conf",
		"/run/systemd/resolve/resolv.conf",
		"/run/systemd/resolve/stub-resolv.conf",
	}

	// WaitCheck is the default wait time before http checks.
	WaitCheck = 1 * time.Second

	// HTTPSTimeout is the default timeout for http requests.
	HTTPSTimeout = 5 * time.Second

	// UntrustedTimer is the default timer for periodic checks in case of
	// an untrusted network.
	UntrustedTimer = 30 * time.Second

	// TrustedTimer is the default timer for periodic checks in case of a
	// trusted network.
	TrustedTimer = 60 * time.Second
)

// Config is a TND configuration.
type Config struct {
	// WatchFiles are the files to watch for changes. By default, they are
	// resolv.conf files in /etc and /run/systemd/resolve.
	WatchFiles []string

	// WaitCheck is the wait time before http checks.
	WaitCheck time.Duration

	// HTTPSTimeout is the timeout for http requests.
	HTTPSTimeout time.Duration

	// UntrustedTimer is the timer for periodic checks in case of an
	// untrusted network.
	UntrustedTimer time.Duration

	// TrustedTimer is the timer for periodic checks in case of a
	// trusted network.
	TrustedTimer time.Duration
}

// Copy returns a copy of Config.
func (c *Config) Copy() *Config {
	tnd := *c
	tnd.WatchFiles = append(c.WatchFiles[:0:0], c.WatchFiles...)

	return &tnd
}

// Valid returns whether Config is valid.
func (c *Config) Valid() bool {
	if c == nil ||
		len(c.WatchFiles) == 0 ||
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
		WatchFiles:     append(WatchFiles[:0:0], WatchFiles...),
		WaitCheck:      WaitCheck,
		HTTPSTimeout:   HTTPSTimeout,
		UntrustedTimer: UntrustedTimer,
		TrustedTimer:   TrustedTimer,
	}
}
