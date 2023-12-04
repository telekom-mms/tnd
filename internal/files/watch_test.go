package files

import (
	"errors"
	"testing"

	"github.com/fsnotify/fsnotify"
)

// TestWatchStartStop tests Start() and Stop() of Watch.
func TestWatchStartStop(t *testing.T) {
	probes := make(chan struct{})

	// error creating fsnotify.Watcher
	t.Run("watcher error", func(t *testing.T) {
		// fail when creating watcher
		defer func() { fsnotifyNewWatcher = fsnotify.NewWatcher }()
		fsnotifyNewWatcher = func() (*fsnotify.Watcher, error) {
			return nil, errors.New("test error")
		}

		// test error
		fw := NewWatch(probes)
		if err := fw.Start(); err == nil {
			t.Errorf("start should fail")
		}
	})

	// error adding dir to watcher
	t.Run("add dir error", func(t *testing.T) {
		// cleanups after test
		oldAdd := watcherAdd
		defer func() { watcherAdd = oldAdd }()

		// test dirs
		for _, dir := range []string{
			etc,
			systemdResolveDir,
		} {
			// fail when adding dir
			watcherAdd = func(_ *fsnotify.Watcher, name string) error {
				if name == dir {
					return errors.New("test error")
				}
				return nil
			}

			// test error
			fw := NewWatch(probes)
			if err := fw.Start(); err == nil {
				t.Errorf("start should fail")
			}
		}
	})

	// no errors
	t.Run("no errors", func(t *testing.T) {
		// cleanups after test
		oldEtc := etc
		oldResolve := systemdResolveDir
		defer func() {
			etc = oldEtc
			systemdResolveDir = oldResolve
		}()

		// create test dirs
		etc = t.TempDir()
		systemdResolveDir = t.TempDir()

		// test without errors
		fw := NewWatch(probes)
		if err := fw.Start(); err != nil {
			t.Errorf("start should not fail: %v", err)
		}
		fw.Stop()
	})
}

// TestNewWatch tests NewWatch.
func TestNewWatch(t *testing.T) {
	probes := make(chan struct{})
	fw := NewWatch(probes)
	if fw.probes != probes {
		t.Errorf("got %p, want %p", fw.probes, probes)
	}
	if fw.done == nil {
		t.Errorf("got nil, want != nil")
	}
	if fw.closed == nil {
		t.Errorf("got nil, want != nil")
	}
}
