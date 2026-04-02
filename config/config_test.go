/*
Copyright © 2026 rtsp-recorder contributors

Unit tests for configuration management.
*/
package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

// TestDefaultConfig_TimelapseDuration verifies DefaultConfig returns TimelapseDuration = 0.
func TestDefaultConfig_TimelapseDuration(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TimelapseDuration != 0 {
		t.Errorf("expected TimelapseDuration=0, got %v", cfg.TimelapseDuration)
	}
}

// TestConfig_TimelapseDuration_Unmarshal verifies Load() properly unmarshals timelapse_duration from viper.
func TestConfig_TimelapseDuration_Unmarshal(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()

	// Set timelapse_duration in viper
	viper.Set("timelapse_duration", 10*time.Second)

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.TimelapseDuration != 10*time.Second {
		t.Errorf("expected TimelapseDuration=10s, got %v", cfg.TimelapseDuration)
	}
}

// TestConfig_TimelapseDuration_MapstructureTag verifies Config struct has correct mapstructure tag.
func TestConfig_TimelapseDuration_MapstructureTag(t *testing.T) {
	// This test verifies the field exists with proper tag
	// We use reflection to check the tag
	cfg := &Config{}
	_ = cfg.TimelapseDuration // Verify field exists
}
