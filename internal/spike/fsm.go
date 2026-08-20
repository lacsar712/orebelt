package spike

import (
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// FSM implements Idle → InSpike → Idle spike detection.
type FSM struct {
	params Params
	state  State

	consecutiveHits  int
	consecutiveMiss  int
	currentPeak      float64
	spikeStart       time.Time
	lastSampleTS     time.Time
	samplesInSpike   int
}

// NewFSM constructs a spike detector with the given parameters.
func NewFSM(p Params) *FSM {
	return &FSM{params: p, state: StateIdle}
}

// State returns the current FSM phase.
func (f *FSM) State() State {
	return f.state
}

// Observe ingests one sample and returns a completed candidate when a spike ends.
func (f *FSM) Observe(s model.Sample) *Candidate {
	hit := s.AbsAccel() > f.params.ThresholdG
	f.lastSampleTS = s.TS

	switch f.state {
	case StateIdle:
		return f.observeIdle(s, hit)
	case StateInSpike:
		return f.observeInSpike(s, hit)
	default:
		f.state = StateIdle
		return nil
	}
}

func (f *FSM) observeIdle(s model.Sample, hit bool) *Candidate {
	if hit {
		if f.consecutiveHits == 0 {
			f.spikeStart = s.TS
			f.currentPeak = s.AbsAccel()
		} else if s.AbsAccel() > f.currentPeak {
			f.currentPeak = s.AbsAccel()
		}
		f.consecutiveHits++
		if f.consecutiveHits >= f.params.MinSamples {
			f.state = StateInSpike
			f.consecutiveMiss = 0
			f.samplesInSpike = f.consecutiveHits
			return nil
		}
	} else {
		f.consecutiveHits = 0
		f.currentPeak = 0
		f.spikeStart = time.Time{}
	}
	return nil
}

func (f *FSM) observeInSpike(s model.Sample, hit bool) *Candidate {
	if hit {
		f.consecutiveMiss = 0
		f.samplesInSpike++
		if s.AbsAccel() > f.currentPeak {
			f.currentPeak = s.AbsAccel()
		}
		return nil
	}

	f.consecutiveMiss++
	if f.consecutiveMiss >= f.params.EndSamples {
		return f.finishSpike(s.TS)
	}
	return nil
}

// Tick handles timeout-based exit when samples stop arriving.
func (f *FSM) Tick(now time.Time) *Candidate {
	if f.state != StateInSpike {
		return nil
	}
	if f.params.Timeout <= 0 {
		return nil
	}
	if now.Sub(f.lastSampleTS) >= f.params.Timeout {
		return f.finishSpike(f.lastSampleTS)
	}
	return nil
}

func (f *FSM) finishSpike(end time.Time) *Candidate {
	c := &Candidate{
		BeltID:    "",
		Peak:      f.currentPeak,
		Start:     f.spikeStart,
		End:       end,
		SampleCnt: f.samplesInSpike,
	}
	f.reset()
	return c
}

func (f *FSM) reset() {
	f.state = StateIdle
	f.consecutiveHits = 0
	f.consecutiveMiss = 0
	f.currentPeak = 0
	f.spikeStart = time.Time{}
	f.samplesInSpike = 0
}

// Reset forces the FSM back to idle without emitting.
func (f *FSM) Reset() {
	f.reset()
}
