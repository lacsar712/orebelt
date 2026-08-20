package emit

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/spike"
	"github.com/lacsar712/orebelt/internal/timeutil"
)

// TestHTTPSinkHonorsCancel reproduces orebelt-005 (STACK-friendly).
func TestHTTPSinkHonorsCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	block := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	})}
	go srv.Serve(ln)
	defer func() {
		close(block)
		_ = srv.Close()
	}()

	sink := NewHTTPSink("http://"+ln.Addr().String()+"/spike", nil)
	store := NewStore(10)
	clock := timeutil.FixedClock{T: time.Now()}
	em := NewEmitter(store, sink, clock, nil)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- em.Emit(ctx, model.SpikeEvent{BeltID: "1", Peak: 3.5, EmittedAt: clock.Now()})
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("STACK: Emit returned nil after cancel; outbound HTTP must honor ctx (not context.Background)")
		}
		if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context") {
			t.Fatalf("STACK: want context cancel error after cancel during HTTP post, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("STACK: Emit hung after cancel — likely NewRequestWithContext(context.Background) or Send(Background)")
	}
}

// TestStoreKeepsNewestOnTrim reproduces orebelt-010.
func TestStoreKeepsNewestOnTrim(t *testing.T) {
	s := NewStore(3)
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		s.Add(model.SpikeEvent{
			BeltID:    "belt-e",
			Peak:      float64(i),
			EmittedAt: base.Add(time.Duration(i) * time.Second),
		})
	}
	if s.Count("belt-e") != 3 {
		t.Fatalf("want trimmed size 3, got %d", s.Count("belt-e"))
	}
	recent := s.Recent("belt-e", 3)
	if len(recent) != 3 {
		t.Fatalf("want 3 recent, got %d", len(recent))
	}
	if recent[0].Peak != 2 || recent[1].Peak != 3 || recent[2].Peak != 4 {
		t.Fatalf("trim must keep newest events, got peaks %+v want [2 3 4]", []float64{recent[0].Peak, recent[1].Peak, recent[2].Peak})
	}
	sum := s.Summary("belt-e")
	if sum.LastPeak != 4 {
		t.Fatalf("Summary LastPeak want 4 (newest), got %.0f", sum.LastPeak)
	}

	exp := s.Export()
	list := exp["belt-e"]
	if len(list) != 3 || list[len(list)-1].Peak != 4 {
		t.Fatalf("Export must retain newest-tail trim: %+v", list)
	}
}

// TestEmitFromCandidatePreservesBeltID is a helper check used with processor path for orebelt-009.
func TestEmitFromCandidatePreservesBeltID(t *testing.T) {
	store := NewStore(5)
	em := NewEmitter(store, NewHTTPSink("", nil), timeutil.FixedClock{T: time.Now()}, nil)
	err := em.EmitFromCandidate(context.Background(), spike.Candidate{
		BeltID: "belt-9", Peak: 4, SampleCnt: 3,
		Start: time.Now(), End: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := store.Recent("belt-9", 1)
	if len(got) != 1 || got[0].BeltID != "belt-9" {
		t.Fatalf("emitted event missing BeltID: %+v", got)
	}
}
