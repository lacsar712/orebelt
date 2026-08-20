package emit

import (
	"sync"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// Store retains emitted spike events for API queries.
type Store struct {
	mu          sync.RWMutex
	maxPerBelt  int
	byBelt      map[string][]model.SpikeEvent
	totalByBelt map[string]int
}

// NewStore creates an in-memory spike event archive.
func NewStore(maxPerBelt int) *Store {
	return &Store{
		maxPerBelt:  maxPerBelt,
		byBelt:      make(map[string][]model.SpikeEvent),
		totalByBelt: make(map[string]int),
	}
}

// Add records a newly emitted spike event.
func (s *Store) Add(ev model.SpikeEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.byBelt[ev.BeltID]
	list = append(list, ev)
	if len(list) > s.maxPerBelt {
		list = list[len(list)-s.maxPerBelt:]
	}
	s.byBelt[ev.BeltID] = list
	s.totalByBelt[ev.BeltID]++
}

// Recent returns the most recent events for a belt, newest last.
func (s *Store) Recent(beltID string, limit int) []model.SpikeEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.byBelt[beltID]
	if limit <= 0 || limit > len(list) {
		limit = len(list)
	}
	if limit == 0 {
		return nil
	}
	start := len(list) - limit
	out := make([]model.SpikeEvent, limit)
	copy(out, list[start:])
	return out
}

// Summary builds dashboard data for one belt.
func (s *Store) Summary(beltID string) model.BeltSpikeSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.byBelt[beltID]
	sum := model.BeltSpikeSummary{
		BeltID:     beltID,
		TotalCount: s.totalByBelt[beltID],
	}
	if len(list) == 0 {
		return sum
	}
	last := list[len(list)-1]
	sum.LastPeak = last.Peak
	sum.LastStart = last.Start
	sum.LastEnd = last.End
	sum.Recent = make([]model.SpikeEvent, len(list))
	copy(sum.Recent, list)
	return sum
}

// BeltIDs returns belts with at least one stored event.
func (s *Store) BeltIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.byBelt))
	for id := range s.byBelt {
		ids = append(ids, id)
	}
	return ids
}

// AllSummaries returns summaries for every belt with history.
func (s *Store) AllSummaries() []model.BeltSpikeSummary {
	ids := s.BeltIDs()
	out := make([]model.BeltSpikeSummary, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.Summary(id))
	}
	return out
}

// Clear removes stored events for a belt.
func (s *Store) Clear(beltID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byBelt, beltID)
	delete(s.totalByBelt, beltID)
}

// Count returns how many events are stored for a belt.
func (s *Store) Count(beltID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byBelt[beltID])
}

// LastEmittedAt returns the timestamp of the newest stored event.
func (s *Store) LastEmittedAt(beltID string) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.byBelt[beltID]
	if len(list) == 0 {
		return time.Time{}
	}
	return list[len(list)-1].EmittedAt
}
