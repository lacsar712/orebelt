package config

import (
	"flag"
	"fmt"
	"time"
)

// Flags registers command-line flags into a Config.
type Flags struct {
	Listen       *string
	RingCap      *int
	Threshold    *float64
	MinSamples   *int
	EndSamples   *int
	SpikeTimeout *time.Duration
	Throttle     *time.Duration
	FlushEvery   *time.Duration
	EmitURL      *string
	LogLevel     *string
}

// Register binds orebelt flags to the given flag set.
func Register(fs *flag.FlagSet) Flags {
	d := Default()
	return Flags{
		Listen:       fs.String("listen", d.ListenAddr, "HTTP listen address"),
		RingCap:      fs.Int("ring-capacity", d.RingCapacity, "Ring buffer capacity per belt"),
		Threshold:    fs.Float64("spike-threshold", d.SpikeThresholdG, "Spike threshold in g"),
		MinSamples:   fs.Int("spike-min-samples", d.SpikeMinSamples, "Consecutive hits to enter spike"),
		EndSamples:   fs.Int("spike-end-samples", d.SpikeEndSamples, "Consecutive misses to exit spike"),
		SpikeTimeout: fs.Duration("spike-timeout", d.SpikeTimeout, "Spike idle timeout"),
		Throttle:     fs.Duration("throttle", d.ThrottleInterval, "Minimum emit interval per belt"),
		FlushEvery:   fs.Duration("flush-interval", d.FlushInterval, "Pending flush ticker interval"),
		EmitURL:      fs.String("emit-url", d.EmitURL, "Optional maintenance webhook URL"),
		LogLevel:     fs.String("log-level", d.LogLevel, "Log level: debug, info, warn, error"),
	}
}

// Apply copies parsed flag values into cfg.
func (f Flags) Apply(cfg *Config) {
	cfg.ListenAddr = *f.Listen
	cfg.RingCapacity = *f.RingCap
	cfg.SpikeThresholdG = *f.Threshold
	cfg.SpikeMinSamples = *f.MinSamples
	cfg.SpikeEndSamples = *f.EndSamples
	cfg.SpikeTimeout = *f.SpikeTimeout
	cfg.ThrottleInterval = *f.Throttle
	cfg.FlushInterval = *f.FlushEvery
	cfg.EmitURL = *f.EmitURL
	cfg.LogLevel = *f.LogLevel
}

// ParseArgs builds configuration from defaults, environment, and CLI flags.
func ParseArgs(args []string) (Config, error) {
	fs := flag.NewFlagSet("orebelt", flag.ContinueOnError)
	fl := Register(fs)
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	cfg := FromEnv(Default())
	fl.Apply(&cfg)
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
