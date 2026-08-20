package ring

import (
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

func TestBufferPushAndSnapshot(t *testing.T) {
	b := New(3)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		b.Push(model.Sample{BeltID: "1", Accel: float64(i), TS: base.Add(time.Duration(i) * time.Second)})
	}
	snap := b.Snapshot("1")
	if len(snap) != 3 {
		t.Fatalf("want 3 samples, got %d", len(snap))
	}
	if snap[0].Accel != 2 || snap[2].Accel != 4 {
		t.Fatalf("unexpected order: %+v", snap)
	}
	// Mutating snapshot must not affect buffer.
	snap[0].Accel = 999
	snap2 := b.Snapshot("1")
	if snap2[0].Accel == 999 {
		t.Fatal("snapshot alias leak")
	}
}

func TestBufferLenAndClear(t *testing.T) {
	b := New(10)
	if b.Len("x") != 0 {
		t.Fatal("expected zero len")
	}
	b.Push(model.Sample{BeltID: "x", Accel: 1, TS: time.Now()})
	if b.Len("x") != 1 {
		t.Fatal("expected len 1")
	}
	b.Clear("x")
	if b.Len("x") != 0 {
		t.Fatal("expected cleared")
	}
}

func TestBufferWindow(t *testing.T) {
	b := New(10)
	base := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		b.Push(model.Sample{BeltID: "w", Accel: 1, TS: base.Add(time.Duration(i) * time.Minute)})
	}
	win := b.Window("w", base.Add(time.Minute), base.Add(2*time.Minute))
	if len(win) != 2 {
		t.Fatalf("want 2 in window, got %d", len(win))
	}
}

func TestCollectStats(t *testing.T) {
	b := New(5)
	ts := time.Now()
	b.Push(model.Sample{BeltID: "a", Accel: 1, TS: ts})
	stats := b.CollectStats()
	if len(stats) != 1 || stats[0].Count != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
