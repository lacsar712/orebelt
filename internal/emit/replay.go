package emit

import (
	"context"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// ReplayOptions controls historical event replay.
type ReplayOptions struct {
	Since   time.Time
	Until   time.Time
	BeltID  string
	Delay   time.Duration
	OnEvent func(model.SpikeEvent) error
}

// Replay walks stored events and invokes onEvent for each match.
func (s *Store) Replay(ctx context.Context, opts ReplayOptions) (int, error) {
	s.mu.RLock()
	var events []model.SpikeEvent
	if opts.BeltID != "" {
		events = append([]model.SpikeEvent(nil), s.byBelt[opts.BeltID]...)
	} else {
		for _, list := range s.byBelt {
			events = append(events, list...)
		}
	}
	s.mu.RUnlock()

	count := 0
	handler := opts.OnEvent
	if handler == nil {
		handler = func(model.SpikeEvent) error { return nil }
	}
	for _, ev := range events {
		if err := ctx.Err(); err != nil {
			return count, err
		}
		if !opts.Since.IsZero() && ev.EmittedAt.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && ev.EmittedAt.After(opts.Until) {
			continue
		}
		if err := handler(ev); err != nil {
			return count, err
		}
		count++
		if opts.Delay > 0 {
			timer := time.NewTimer(opts.Delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return count, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return count, nil
}

// Export returns a deep copy of all stored events grouped by belt.
func (s *Store) Export() map[string][]model.SpikeEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]model.SpikeEvent, len(s.byBelt))
	for id, list := range s.byBelt {
		cp := make([]model.SpikeEvent, len(list))
		copy(cp, list)
		out[id] = cp
	}
	return out
}
