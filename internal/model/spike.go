package model

import "time"

// SpikeEvent is an aggregated vibration spike emitted to maintenance systems.
type SpikeEvent struct {
	BeltID    string    `json:"belt_id"`
	Peak      float64   `json:"peak"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	SampleCnt int       `json:"sample_count"`
	EmittedAt time.Time `json:"emitted_at"`
}

// Duration returns the elapsed wall time covered by the spike interval.
func (e SpikeEvent) Duration() time.Duration {
	if e.End.Before(e.Start) {
		return 0
	}
	return e.End.Sub(e.Start)
}

// BeltSpikeSummary summarizes recent spike activity for a belt dashboard row.
type BeltSpikeSummary struct {
	BeltID     string     `json:"belt_id"`
	LastPeak   float64    `json:"last_peak"`
	LastStart  time.Time  `json:"last_start"`
	LastEnd    time.Time  `json:"last_end"`
	TotalCount int        `json:"total_count"`
	Recent     []SpikeEvent `json:"recent"`
}
