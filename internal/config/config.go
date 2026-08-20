package config

import (
	"fmt"
	"time"
)

// Config holds runtime configuration for the orebelt service.
type Config struct {
	ListenAddr       string        `json:"listen_addr"`
	RingCapacity     int           `json:"ring_capacity"`
	SpikeThresholdG  float64       `json:"spike_threshold_g"`
	SpikeMinSamples  int           `json:"spike_min_samples"`
	SpikeEndSamples  int           `json:"spike_end_samples"`
	SpikeTimeout     time.Duration `json:"spike_timeout"`
	ThrottleInterval time.Duration `json:"throttle_interval"`
	FlushInterval    time.Duration `json:"flush_interval"`
	EmitURL          string        `json:"emit_url"`
	MaxRecentSpikes  int           `json:"max_recent_spikes"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	LogLevel         string        `json:"log_level"`
}

// ApplyDefaults fills zero values with production-safe defaults.
func (c *Config) ApplyDefaults() {
	d := Default()
	if c.ListenAddr == "" {
		c.ListenAddr = d.ListenAddr
	}
	if c.RingCapacity <= 0 {
		c.RingCapacity = d.RingCapacity
	}
	if c.SpikeThresholdG <= 0 {
		c.SpikeThresholdG = d.SpikeThresholdG
	}
	if c.SpikeMinSamples <= 0 {
		c.SpikeMinSamples = d.SpikeMinSamples
	}
	if c.SpikeEndSamples <= 0 {
		c.SpikeEndSamples = d.SpikeEndSamples
	}
	if c.SpikeTimeout <= 0 {
		c.SpikeTimeout = d.SpikeTimeout
	}
	if c.ThrottleInterval <= 0 {
		c.ThrottleInterval = d.ThrottleInterval
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = d.FlushInterval
	}
	if c.MaxRecentSpikes <= 0 {
		c.MaxRecentSpikes = d.MaxRecentSpikes
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = d.ReadTimeout
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = d.WriteTimeout
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = d.IdleTimeout
	}
	if c.LogLevel == "" {
		c.LogLevel = d.LogLevel
	}
}

// Validate returns an error when configuration values are inconsistent.
func (c *Config) Validate() error {
	if c.RingCapacity < 16 {
		return fmt.Errorf("ring_capacity must be >= 16, got %d", c.RingCapacity)
	}
	if c.SpikeMinSamples < 1 {
		return fmt.Errorf("spike_min_samples must be >= 1")
	}
	if c.SpikeEndSamples < 1 {
		return fmt.Errorf("spike_end_samples must be >= 1")
	}
	if c.SpikeThresholdG <= 0 {
		return fmt.Errorf("spike_threshold_g must be positive")
	}
	if c.ThrottleInterval < time.Second {
		return fmt.Errorf("throttle_interval must be >= 1s")
	}
	if c.FlushInterval <= 0 {
		return fmt.Errorf("flush_interval must be positive")
	}
	if c.ListenAddr == "" {
		return fmt.Errorf("listen_addr is required")
	}
	return nil
}
