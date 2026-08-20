package emit

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/spike"
	"github.com/lacsar712/orebelt/internal/timeutil"
)

func TestStoreAddAndRecent(t *testing.T) {
	s := NewStore(3)
	now := time.Now()
	for i := 0; i < 5; i++ {
		s.Add(model.SpikeEvent{BeltID: "1", Peak: float64(i), EmittedAt: now})
	}
	if s.Count("1") != 3 {
		t.Fatalf("want trimmed store size 3, got %d", s.Count("1"))
	}
	recent := s.Recent("1", 2)
	if len(recent) != 2 || recent[1].Peak != 4 {
		t.Fatalf("unexpected recent: %+v", recent)
	}
}

func TestEmitterRespectsContext(t *testing.T) {
	store := NewStore(10)
	sink := NewHTTPSink("http://127.0.0.1:1/unreachable", nil)
	clock := timeutil.FixedClock{T: time.Now()}
	e := NewEmitter(store, sink, clock, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := e.EmitFromCandidate(ctx, spike.Candidate{BeltID: "1", Peak: 3})
	if err == nil {
		t.Fatal("expected context error before remote call")
	}
}

func TestHTTPSinkDisabled(t *testing.T) {
	s := NewHTTPSink("", nil)
	if s.Enabled() {
		t.Fatal("empty url should disable sink")
	}
	if err := s.Send(context.Background(), model.SpikeEvent{}); err != nil {
		t.Fatal(err)
	}
}
