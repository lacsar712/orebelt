package ring

import (
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// Stats summarizes ring utilization for observability endpoints.
type Stats struct {
	BeltID     string `json:"belt_id"`
	Count      int    `json:"count"`
	Capacity   int    `json:"capacity"`
	OldestTS   time.Time `json:"oldest_ts,omitempty"`
	NewestTS   time.Time `json:"newest_ts,omitempty"`
}

// CollectStats builds per-belt statistics from the buffer contents.
func (b *Buffer) CollectStats() []Stats {
	ids := b.BeltIDs()
	out := make([]Stats, 0, len(ids))
	for _, id := range ids {
		snap := b.Snapshot(id)
		st := Stats{
			BeltID:   id,
			Count:    len(snap),
			Capacity: b.capacity,
		}
		// Snapshot is time-ascending: snap[0] is the oldest, the last is newest.
		if len(snap) > 0 {
			st.OldestTS = snap[0].TS
			st.NewestTS = snap[len(snap)-1].TS
		}
		out = append(out, st)
	}
	return out
}

// Window returns samples whose timestamps fall within [since, until].
func (b *Buffer) Window(beltID string, since, until time.Time) []model.Sample {
	snap := b.Snapshot(beltID)
	if len(snap) == 0 {
		return nil
	}
	out := make([]model.Sample, 0)
	for _, s := range snap {
		if (since.IsZero() || !s.TS.Before(since)) && (until.IsZero() || !s.TS.After(until)) {
			out = append(out, s)
		}
	}
	return out
}
