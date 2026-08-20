package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// FromEnv overlays environment variables onto a base configuration.
func FromEnv(base Config) Config {
	base.ApplyDefaults()
	if v := os.Getenv("OREBELT_LISTEN"); v != "" {
		base.ListenAddr = v
	}
	if v := os.Getenv("OREBELT_RING_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			base.RingCapacity = n
		}
	}
	if v := os.Getenv("OREBELT_SPIKE_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			base.SpikeThresholdG = f
		}
	}
	if v := os.Getenv("OREBELT_THROTTLE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			base.ThrottleInterval = d
		}
	}
	if v := os.Getenv("OREBELT_EMIT_URL"); v != "" {
		base.EmitURL = v
	}
	if v := os.Getenv("OREBELT_LOG_LEVEL"); v != "" {
		base.LogLevel = strings.ToLower(v)
	}
	return base
}
