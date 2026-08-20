package ring

import (
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// TestSnapshotChronologicalAfterWrap reproduces orebelt-001.
func TestSnapshotChronologicalAfterWrap(t *testing.T) {
	// Non-full ring: correct start is 0, but br.head == count (>0).
	b := New(5)
	base := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		b.Push(model.Sample{
			BeltID: "belt-a",
			Accel:  float64(i + 10),
			TS:     base.Add(time.Duration(i) * time.Second),
		})
	}
	snap := b.Snapshot("belt-a")
	if len(snap) != 3 {
		t.Fatalf("want 3 samples, got %d", len(snap))
	}
	wantAccel := []float64{10, 11, 12}
	wantTS := []time.Time{
		base,
		base.Add(1 * time.Second),
		base.Add(2 * time.Second),
	}
	for i := range wantAccel {
		if snap[i].Accel != wantAccel[i] || !snap[i].TS.Equal(wantTS[i]) {
			t.Fatalf("snapshot not chronological at %d: got accel=%.0f ts=%s want accel=%.0f ts=%s (full=%+v)",
				i, snap[i].Accel, snap[i].TS, wantAccel[i], wantTS[i], snap)
		}
	}

	// After wrap, still chronological (oldest→newest).
	b2 := New(3)
	for i := 0; i < 5; i++ {
		b2.Push(model.Sample{
			BeltID: "belt-w",
			Accel:  float64(i + 20),
			TS:     base.Add(time.Duration(i) * time.Second),
		})
	}
	wrap := b2.Snapshot("belt-w")
	if len(wrap) != 3 || wrap[0].Accel != 22 || wrap[1].Accel != 23 || wrap[2].Accel != 24 {
		t.Fatalf("wrapped snapshot not chronological: %+v", wrap)
	}

	stats := b.CollectStats()
	if len(stats) != 1 {
		t.Fatalf("want 1 stats row, got %d", len(stats))
	}
	if !stats[0].OldestTS.Equal(wantTS[0]) || !stats[0].NewestTS.Equal(wantTS[2]) {
		t.Fatalf("CollectStats timestamps wrong: oldest=%s newest=%s want oldest=%s newest=%s",
			stats[0].OldestTS, stats[0].NewestTS, wantTS[0], wantTS[2])
	}
}

// TestSnapshotIndependentCopy reproduces orebelt-008.
func TestSnapshotIndependentCopy(t *testing.T) {
	b := New(4)
	base := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	b.Push(model.Sample{BeltID: "belt-b", Accel: 1.5, TS: base})
	b.Push(model.Sample{BeltID: "belt-b", Accel: 2.5, TS: base.Add(time.Second)})

	snap := b.Snapshot("belt-b")
	if len(snap) != 2 {
		t.Fatalf("want 2 samples, got %d", len(snap))
	}
	snap[0].Accel = 999
	snap[1].Accel = 888

	again := b.Snapshot("belt-b")
	if again[0].Accel == 999 || again[1].Accel == 888 {
		t.Fatalf("snapshot aliased internal storage: first=%+v second=%+v", again[0], again[1])
	}

	walk := b.Snapshot("belt-b")
	it := NewIterator(walk)
	walk[0].Accel = 777
	got, ok := it.Next()
	if !ok {
		t.Fatal("iterator empty")
	}
	if got.Accel != 1.5 {
		t.Fatalf("iterator must copy inputs (not alias caller slice): got %.1f want 1.5", got.Accel)
	}
}
