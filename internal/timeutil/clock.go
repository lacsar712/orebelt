package timeutil

import "time"

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// RealClock uses the system clock.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// FixedClock returns a constant timestamp.
type FixedClock struct {
	T time.Time
}

func (c FixedClock) Now() time.Time { return c.T }

// Advance moves a fixed clock forward.
func (c *FixedClock) Advance(d time.Duration) {
	c.T = c.T.Add(d)
}
