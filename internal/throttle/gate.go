package throttle

import (
	"sync"
	"time"

	"github.com/lacsar712/orebelt/internal/spike"
)

// Gate limits how frequently spike events may be emitted per belt.
type Gate struct {
	mu       sync.Mutex
	interval time.Duration
	lastEmit map[string]time.Time
	pending  map[string]*pendingEvent
}

type pendingEvent struct {
	candidate spike.Candidate
	queuedAt  time.Time
}

// NewGate constructs a throttle with the given minimum emit interval.
func NewGate(interval time.Duration) *Gate {
	return &Gate{
		interval: interval,
		lastEmit: make(map[string]time.Time),
		pending:  make(map[string]*pendingEvent),
	}
}

// Allow reports whether an event can be emitted immediately.
// When false, the candidate is stored as pending and may be refreshed.
func (g *Gate) Allow(beltID string, now time.Time, c spike.Candidate) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	last, seen := g.lastEmit[beltID]
	if seen && now.Sub(last) < g.interval {
		g.storePending(beltID, now, c)
		return false
	}
	g.lastEmit[beltID] = now
	delete(g.pending, beltID)
	return true
}

func (g *Gate) storePending(beltID string, now time.Time, c spike.Candidate) {
	if p, ok := g.pending[beltID]; ok {
		// Protect the higher-peak pending candidate: a later, lower-peak
		// spike must not overwrite one already queued, otherwise the
		// vibration that actually warrants reporting gets buried. Only
		// refresh when the new candidate is strictly more severe.
		if c.Peak > p.candidate.Peak {
			p.candidate = c
			p.queuedAt = now
		}
		return
	}
	g.pending[beltID] = &pendingEvent{candidate: c, queuedAt: now}
}

// Pending returns a copy of all queued events.
func (g *Gate) Pending() map[string]spike.Candidate {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make(map[string]spike.Candidate, len(g.pending))
	for id, p := range g.pending {
		out[id] = p.candidate
	}
	return out
}

// FlushDue emits pending events whose throttle window has elapsed.
func (g *Gate) FlushDue(now time.Time) []spike.Candidate {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []spike.Candidate
	for beltID, p := range g.pending {
		last := g.lastEmit[beltID]
		if now.Sub(last) >= g.interval {
			out = append(out, p.candidate)
			g.lastEmit[beltID] = now
			delete(g.pending, beltID)
		}
	}
	return out
}

// MarkEmitted records an emission timestamp without going through Allow.
func (g *Gate) MarkEmitted(beltID string, now time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lastEmit[beltID] = now
}

// Interval returns the configured throttle interval.
func (g *Gate) Interval() time.Duration {
	return g.interval
}
