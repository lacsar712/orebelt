package spike

import (
	"sync"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// Detector manages one FSM per belt identifier.
type Detector struct {
	mu     sync.Mutex
	params Params
	belts  map[string]*FSM
}

// NewDetector creates a multi-belt spike detector.
func NewDetector(p Params) *Detector {
	return &Detector{
		params: p,
		belts:  make(map[string]*FSM),
	}
}

func (d *Detector) fsm(beltID string) *FSM {
	f, ok := d.belts[beltID]
	if !ok {
		f = NewFSM(d.params)
		d.belts[beltID] = f
	}
	return f
}

// Observe routes a sample to the belt-specific FSM.
func (d *Detector) Observe(s model.Sample) *Candidate {
	d.mu.Lock()
	defer d.mu.Unlock()
	c := d.fsm(s.BeltID).Observe(s)
	if c != nil {
		c.BeltID = s.BeltID
	}
	return c
}

// Tick runs timeout checks for every belt FSM.
func (d *Detector) Tick(now time.Time) []Candidate {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []Candidate
	for beltID, f := range d.belts {
		if c := f.Tick(now); c != nil {
			c.BeltID = beltID
			out = append(out, *c)
		}
	}
	return out
}

// BeltState returns diagnostic state for a belt.
func (d *Detector) BeltState(beltID string) State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.fsm(beltID).State()
}

// ResetBelt clears detection state for one belt.
func (d *Detector) ResetBelt(beltID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if f, ok := d.belts[beltID]; ok {
		f.Reset()
	}
}

// Params returns a copy of detector parameters.
func (d *Detector) Params() Params {
	return d.params
}
