package files

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/fsnotify/fsnotify"
)

// testFiles are resolv.conf files for testing.
var testFiles = []string{
	"/etc/resolv.conf",
	"/run/systemd/resolve/resolv.conf",
	"/run/systemd/resolve/stub-resolv.conf",
}

// TestWatchStartEvents tests start of Watch, events.
func TestWatchStartEvents(t *testing.T) {
	// create watcher
	probes := make(chan struct{})
	fw := NewWatch(probes, testFiles)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	fw.watcher = w

	// start watcher and get initial probe
	go fw.start()
	<-probes

	// send watcher events, handle probes
	fw.watcher.Errors <- errors.New("test error")
	fw.watcher.Events <- fsnotify.Event{Name: "/etc/resolv.conf"}
	<-probes

	// unexpected close of watcher channels
	if err := fw.watcher.Close(); err != nil {
		t.Errorf("error closing watcher: %v", err)
	}

	// wait for watcher
	<-fw.closed
}

// TestWatchStartStop tests Start and Stop of Watch.
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
		fw := NewWatch(probes, testFiles)
		if err := fw.Start(); err == nil {
			t.Errorf("start should fail")
		}
	})

	// error adding dir to watcher
	t.Run("add dir error", func(t *testing.T) {
		// cleanups after test
		oldAdd := watcherAdd
		defer func() { watcherAdd = oldAdd }()

		// fail when adding dir
		watcherAdd = func(_ *fsnotify.Watcher, name string) error {
			return errors.New("test error")
		}

		// test error
		fw := NewWatch(probes, testFiles)
		if err := fw.Start(); err == nil {
			t.Errorf("start should fail")
		}
	})

	// no errors
	t.Run("no errors", func(t *testing.T) {
		// create test dir
		dir := t.TempDir()
		file := filepath.Join(dir, "resolv.conf")

		// test without errors
		fw := NewWatch(probes, []string{file})
		if err := fw.Start(); err != nil {
			t.Errorf("start should not fail: %v", err)
		}
		fw.Stop()
	})
}

// TestNewWatch tests NewWatch.
func TestNewWatch(t *testing.T) {
	probes := make(chan struct{})
	fw := NewWatch(probes, testFiles)
	if !reflect.DeepEqual(fw.files, testFiles) {
		t.Errorf("got %v, want %v", fw.files, testFiles)
	}
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
