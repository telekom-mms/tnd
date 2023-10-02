package tnd

import "testing"

// TestConfigValid tests Valid of Config.
func TestConfigValid(t *testing.T) {
	// test invalid
	for _, invalid := range []*Config{
		nil,
		{-1, 99, 99, 99},
		{99, -1, 99, 99},
		{99, 99, -1, 99},
		{99, 99, 99, -1},
	} {
		if invalid.Valid() {
			t.Errorf("Config should be invalid: %v", invalid)
		}
	}

	// test valid
	for _, valid := range []*Config{
		NewConfig(),
		{1000000000, 1000000000, 1000000000, 1000000000},
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
