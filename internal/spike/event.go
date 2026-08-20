package spike

import "time"

// State represents the spike detection finite-state machine phase.
type State int

const (
	StateIdle State = iota
	StateInSpike
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateInSpike:
		return "in_spike"
	default:
		return "unknown"
	}
}

// Candidate holds a completed spike interval before throttling.
type Candidate struct {
	BeltID     string
	Peak       float64
	Start      time.Time
	End        time.Time
	SampleCnt  int
}

// Params configures spike detection sensitivity.
type Params struct {
	ThresholdG  float64
	MinSamples  int
	EndSamples  int
	Timeout     time.Duration
}
