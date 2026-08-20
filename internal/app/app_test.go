package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/config"
	"github.com/lacsar712/orebelt/internal/emit"
	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/timeutil"
)

func testConfig() config.Config {
	cfg := config.Default()
	cfg.ThrottleInterval = 30 * time.Second
	cfg.SpikeMinSamples = 3
	cfg.SpikeEndSamples = 2
	cfg.SpikeThresholdG = 2.5
	return cfg
}

func TestProcessorSpikeAndThrottle(t *testing.T) {
	cfg := testConfig()
	clock := &timeutil.FixedClock{T: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)}
	store := emit.NewStore(10)
	emitter := emit.NewEmitter(store, emit.NewHTTPSink("", nil), clock, logx.New())
	reg := NewRegistry()
	p := NewProcessor(cfg, emitter, reg, clock, nil)
	ctx := context.Background()

	push := func(accel float64, sec int) {
		ts := clock.Now().Add(time.Duration(sec) * time.Millisecond)
		_ = p.Process(ctx, model.Sample{BeltID: "3", Accel: accel, TS: ts})
	}

	push(3, 0)
	push(3, 1)
	push(3, 2)
	push(3, 3)
	push(0, 10)
	push(0, 11)
	if store.Count("3") != 1 {
		t.Fatalf("expected one emitted spike, got %d", store.Count("3"))
	}
	push(3, 12)
	push(3, 13)
	push(3, 14)
	push(0, 20)
	push(0, 21)
	if store.Count("3") != 1 {
		t.Fatalf("second spike should be throttled, count=%d", store.Count("3"))
	}
	if len(p.Gate().Pending()) != 1 {
		t.Fatal("expected pending spike")
	}
}

func TestAppHandler(t *testing.T) {
	cfg := testConfig()
	cfg.ListenAddr = ":0"
	a, err := New(cfg, logx.New())
	if err != nil {
		t.Fatal(err)
	}
	if a.Handler() == nil {
		t.Fatal("expected handler")
	}
}

func TestRegistryEnsure(t *testing.T) {
	r := NewRegistry()
	r.Ensure("5")
	meta := r.Get("5")
	if meta.ID != "5" {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}
