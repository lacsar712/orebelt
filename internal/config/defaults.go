package config

import "time"

// Default returns the baseline configuration described in BUSINESS.md.
func Default() Config {
	return Config{
		ListenAddr:       ":8080",
		RingCapacity:     4096,
		SpikeThresholdG:  2.5,
		SpikeMinSamples:  3,
		SpikeEndSamples:  2,
		SpikeTimeout:     500 * time.Millisecond,
		ThrottleInterval: 30 * time.Second,
		FlushInterval:    5 * time.Second,
		EmitURL:          "",
		MaxRecentSpikes:  50,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     10 * time.Second,
		IdleTimeout:      120 * time.Second,
		LogLevel:         "info",
	}
}
