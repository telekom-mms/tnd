package tnd

import (
	"reflect"
	"testing"
	"time"
)

// TestConfigCopy tests Copy of Config.
func TestConfigCopy(t *testing.T) {
	// test with new config
	want := NewConfig()
	got := want.Copy()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// test modification after copy
	c1 := NewConfig()
	c2 := c1.Copy()

	c1.WatchFiles[0] = "something else"
	c1.HTTPSTimeout = 3 * time.Second

	if reflect.DeepEqual(c1, c2) {
		t.Error("copies should not be equal after modification")
	}
}

// TestConfigValid tests Valid of Config.
func TestConfigValid(t *testing.T) {
	// test invalid
	for _, invalid := range []*Config{
		nil,
		{nil, 99, 99, 99, 99},
		{[]string{}, 99, 99, 99, 99},
		{WatchFiles, -1, 99, 99, 99},
		{WatchFiles, 99, -1, 99, 99},
		{WatchFiles, 99, 99, -1, 99},
		{WatchFiles, 99, 99, 99, -1},
	} {
		if invalid.Valid() {
			t.Errorf("Config should be invalid: %v", invalid)
		}
	}

	// test valid
	for _, valid := range []*Config{
		NewConfig(),
		{WatchFiles, 1000000000, 1000000000, 1000000000, 1000000000},
	} {
		if !valid.Valid() {
			t.Errorf("Config should be valid: %v", valid)
		}
	}
}

// TestNewConfig tests NewConfig.
func TestNewConfig(t *testing.T) {
	c := NewConfig()
	if !c.Valid() {
		t.Errorf("New config should be valid: %v", c)
	}
}
