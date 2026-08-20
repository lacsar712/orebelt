package ring

import (
	"sort"
	"sync"

	"github.com/lacsar712/orebelt/internal/model"
)

// Buffer is a fixed-capacity ring storing the most recent samples per belt.
type Buffer struct {
	mu       sync.RWMutex
	capacity int
	// belts maps belt ID to per-belt ring storage.
	belts map[string]*beltRing
}

type beltRing struct {
	samples []model.Sample
	head    int
	count   int
}

// New creates a ring buffer with the given per-belt capacity.
func New(capacity int) *Buffer {
	return &Buffer{
		capacity: capacity,
		belts:    make(map[string]*beltRing),
	}
}

// Push appends a sample, overwriting the oldest entry when full.
func (b *Buffer) Push(s model.Sample) {
	b.mu.Lock()
	defer b.mu.Unlock()
	br, ok := b.belts[s.BeltID]
	if !ok {
		br = &beltRing{samples: make([]model.Sample, b.capacity)}
		b.belts[s.BeltID] = br
	}
	br.samples[br.head] = s
	br.head = (br.head + 1) % b.capacity
	if br.count < b.capacity {
		br.count++
	}
}

// Len returns how many samples are stored for a belt.
func (b *Buffer) Len(beltID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	br, ok := b.belts[beltID]
	if !ok {
		return 0
	}
	return br.count
}

// BeltIDs returns the set of belt identifiers currently tracked.
func (b *Buffer) BeltIDs() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	ids := make([]string, 0, len(b.belts))
	for id := range b.belts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Snapshot returns a time-ascending copy of stored samples for a belt.
// The returned slice is independent from internal storage to prevent alias leaks.
func (b *Buffer) Snapshot(beltID string) []model.Sample {
	b.mu.RLock()
	defer b.mu.RUnlock()
	br, ok := b.belts[beltID]
	if !ok || br.count == 0 {
		return nil
	}
	return br.samples[:br.count]
}

// Clear removes all samples for a belt.
func (b *Buffer) Clear(beltID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.belts, beltID)
}

// Capacity returns the configured ring capacity.
func (b *Buffer) Capacity() int {
	return b.capacity
}
