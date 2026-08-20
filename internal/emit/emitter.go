package emit

import (
	"context"
	"sync"

	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/spike"
	"github.com/lacsar712/orebelt/internal/timeutil"
)

// Emitter coordinates local storage and optional remote forwarding.
type Emitter struct {
	mu     sync.Mutex
	store  *Store
	sink   *HTTPSink
	clock  timeutil.Clock
	logger *logx.Logger
}

// NewEmitter constructs the event emission pipeline.
func NewEmitter(store *Store, sink *HTTPSink, clock timeutil.Clock, logger *logx.Logger) *Emitter {
	if clock == nil {
		clock = timeutil.RealClock{}
	}
	return &Emitter{
		store:  store,
		sink:   sink,
		clock:  clock,
		logger: logger,
	}
}

// EmitFromCandidate converts and publishes a spike candidate.
func (e *Emitter) EmitFromCandidate(ctx context.Context, c spike.Candidate) error {
	ev := model.SpikeEvent{
		BeltID:    c.BeltID,
		Peak:      c.Peak,
		Start:     c.Start,
		End:       c.End,
		SampleCnt: c.SampleCnt,
		EmittedAt: e.clock.Now(),
	}
	return e.Emit(ctx, ev)
}

// Emit stores and optionally forwards a spike event.
func (e *Emitter) Emit(ctx context.Context, ev model.SpikeEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	e.store.Add(ev)
	e.mu.Unlock()

	if e.sink == nil || !e.sink.Enabled() {
		return nil
	}
	if err := e.sink.Send(context.Background(), ev); err != nil {
		if e.logger != nil {
			e.logger.Error("remote emit failed belt=%s peak=%.3f: %v", ev.BeltID, ev.Peak, err)
		}
		return err
	}
	if e.logger != nil {
		e.logger.Info("emitted spike belt=%s peak=%.3fg", ev.BeltID, ev.Peak)
	}
	return nil
}

// EmitBatch publishes multiple candidates sequentially, stopping on context cancel.
func (e *Emitter) EmitBatch(ctx context.Context, candidates []spike.Candidate) error {
	for _, c := range candidates {
		if err := e.EmitFromCandidate(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// Store returns the backing event store.
func (e *Emitter) Store() *Store {
	return e.store
}

// Shutdown waits for in-flight work. Currently a no-op placeholder for extension.
func (e *Emitter) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
