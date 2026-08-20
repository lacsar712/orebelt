package model

import "time"

// Sample represents a single belt acceleration reading from an edge collector.
type Sample struct {
	BeltID string    `json:"belt_id"`
	Accel  float64   `json:"accel"`
	TS     time.Time `json:"ts"`
}

// AbsAccel returns the absolute acceleration magnitude in g-units.
func (s Sample) AbsAccel() float64 {
	if s.Accel < 0 {
		return -s.Accel
	}
	return s.Accel
}

// Valid checks whether the sample carries the minimum required fields.
func (s Sample) Valid() bool {
	return s.BeltID != "" && !s.TS.IsZero()
}
