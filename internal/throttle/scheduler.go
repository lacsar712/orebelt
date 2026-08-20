package throttle

import (
	"sync"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/spike"
)

// Scheduler periodically flushes pending spike events.
type Scheduler struct {
	mu       sync.Mutex
	gate     *Gate
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	onFlush  func([]spike.Candidate)
}

// NewScheduler wires a flush ticker to a throttle gate.
func NewScheduler(gate *Gate, interval time.Duration, onFlush func([]spike.Candidate)) *Scheduler {
	return &Scheduler{
		gate:     gate,
		interval: interval,
		onFlush:  onFlush,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start launches the background flush loop.
func (s *Scheduler) Start(now func() time.Time) {
	go func() {
		defer close(s.doneCh)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.flush(now())
			case <-s.stopCh:
				s.flush(now())
				return
			}
		}
	}()
}

func (s *Scheduler) flush(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.onFlush == nil {
		return
	}
	due := s.gate.FlushDue(now)
	if len(due) > 0 {
		s.onFlush(due)
	}
}

// Stop terminates the scheduler and performs a final flush.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// Stats exposes throttle diagnostics for dashboards.
type Stats struct {
	PendingCount int                    `json:"pending_count"`
	Pending      map[string]model.SpikeEvent `json:"pending"`
}

// CollectStats builds a snapshot of pending throttle state.
func CollectStats(gate *Gate) Stats {
	pending := gate.Pending()
	st := Stats{PendingCount: len(pending), Pending: make(map[string]model.SpikeEvent, len(pending))}
	for id, c := range pending {
		st.Pending[id] = model.SpikeEvent{
			BeltID:    c.BeltID,
			Peak:      c.Peak,
			Start:     c.Start,
			End:       c.End,
			SampleCnt: c.SampleCnt,
		}
	}
	return st
}
