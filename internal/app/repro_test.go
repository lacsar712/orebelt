package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/emit"
	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/timeutil"
)

// TestTickTimeoutsFinalizesSpike reproduces orebelt-007 (STACK-friendly).
func TestTickTimeoutsFinalizesSpike(t *testing.T) {
	cfg := testConfig()
	cfg.SpikeMinSamples = 2
	cfg.SpikeEndSamples = 10
	cfg.SpikeTimeout = 100 * time.Millisecond
	cfg.ThrottleInterval = 0

	clock := &timeutil.FixedClock{T: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)}
	store := emit.NewStore(10)
	emitter := emit.NewEmitter(store, emit.NewHTTPSink("", nil), clock, logx.New())
	p := NewProcessor(cfg, emitter, NewRegistry(), clock, nil)
	ctx := context.Background()

	_ = p.Process(ctx, model.Sample{BeltID: "belt-7", Accel: 3.0, TS: clock.Now()})
	_ = p.Process(ctx, model.Sample{BeltID: "belt-7", Accel: 3.1, TS: clock.Now().Add(time.Millisecond)})
	if p.Detector().BeltState("belt-7").String() != "in_spike" {
		t.Fatalf("STACK: setup failed — detector should be in_spike before timeout tick, state=%s", p.Detector().BeltState("belt-7"))
	}

	clock.Advance(200 * time.Millisecond)
	if err := p.TickTimeouts(ctx); err != nil {
		t.Fatalf("STACK: TickTimeouts returned error: %v", err)
	}
	if store.Count("belt-7") != 1 {
		t.Fatalf("STACK: TickTimeouts must call detector.Tick (not Observe) to finalize timed-out spikes; emitted=%d state=%s",
			store.Count("belt-7"), p.Detector().BeltState("belt-7"))
	}
}

// TestProcessorEmitsBeltID reproduces orebelt-009 via processor path.
func TestProcessorEmitsBeltID(t *testing.T) {
	cfg := testConfig()
	cfg.SpikeMinSamples = 2
	cfg.SpikeEndSamples = 1
	cfg.ThrottleInterval = 0

	clock := &timeutil.FixedClock{T: time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)}
	store := emit.NewStore(10)
	emitter := emit.NewEmitter(store, emit.NewHTTPSink("", nil), clock, nil)
	p := NewProcessor(cfg, emitter, NewRegistry(), clock, nil)
	ctx := context.Background()

	_ = p.Process(ctx, model.Sample{BeltID: "belt-9", Accel: 3.0, TS: clock.Now()})
	_ = p.Process(ctx, model.Sample{BeltID: "belt-9", Accel: 3.2, TS: clock.Now().Add(time.Millisecond)})
	_ = p.Process(ctx, model.Sample{BeltID: "belt-9", Accel: 0.1, TS: clock.Now().Add(2 * time.Millisecond)})

	got := store.Recent("belt-9", 1)
	if len(got) != 1 {
		t.Fatalf("expected one emitted spike for belt-9, got %d (all belts=%v)", store.Count("belt-9"), store.BeltIDs())
	}
	if got[0].BeltID != "belt-9" {
		t.Fatalf("processor emit must preserve candidate BeltID, got %q", got[0].BeltID)
	}
}
