package app

import (
	"context"
	"sync"

	"github.com/lacsar712/orebelt/internal/config"
	"github.com/lacsar712/orebelt/internal/emit"
	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/ring"
	"github.com/lacsar712/orebelt/internal/spike"
	"github.com/lacsar712/orebelt/internal/throttle"
	"github.com/lacsar712/orebelt/internal/timeutil"
)

// Processor implements ingest.Processor by running ring, spike, throttle, emit.
type Processor struct {
	mu        sync.Mutex
	cfg       config.Config
	buffer    *ring.Buffer
	detector  *spike.Detector
	gate      *throttle.Gate
	emitter   *emit.Emitter
	registry  *Registry
	clock     timeutil.Clock
	logger    *logx.Logger
}

// NewProcessor constructs the core sample processing pipeline.
func NewProcessor(cfg config.Config, emitter *emit.Emitter, registry *Registry, clock timeutil.Clock, logger *logx.Logger) *Processor {
	if clock == nil {
		clock = timeutil.RealClock{}
	}
	params := spike.Params{
		ThresholdG: cfg.SpikeThresholdG,
		MinSamples: cfg.SpikeMinSamples,
		EndSamples: cfg.SpikeEndSamples,
		Timeout:    cfg.SpikeTimeout,
	}
	return &Processor{
		cfg:      cfg,
		buffer:   ring.New(cfg.RingCapacity),
		detector: spike.NewDetector(params),
		gate:     throttle.NewGate(cfg.ThrottleInterval),
		emitter:  emitter,
		registry: registry,
		clock:    clock,
		logger:   logger,
	}
}

// Process handles one sample end-to-end.
func (p *Processor) Process(ctx context.Context, s model.Sample) error {
	return p.ProcessBatch(ctx, []model.Sample{s})
}

// ProcessBatch ingests an ordered list of samples.
func (p *Processor) ProcessBatch(ctx context.Context, samples []model.Sample) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range samples {
		if err := ctx.Err(); err != nil {
			return err
		}
		p.registry.Ensure(s.BeltID)
		p.buffer.Push(s)
		if c := p.detector.Observe(s); c != nil {
			if err := p.tryEmit(ctx, *c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Processor) tryEmit(ctx context.Context, c spike.Candidate) error {
	now := p.clock.Now()
	if p.gate.Allow(c.BeltID, now, c) {
		// The gate has selected this belt for immediate emission; preserve the
		// candidate's BeltID so the platform can archive the event per belt.
		return p.emitter.EmitFromCandidate(ctx, c)
	}
	if p.logger != nil {
		p.logger.Debug("throttled belt=%s peak=%.3f pending refresh", c.BeltID, c.Peak)
	}
	return nil
}

// FlushPending emits throttle-eligible pending events.
func (p *Processor) FlushPending(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.clock.Now()
	due := p.gate.FlushDue(now)
	for _, c := range due {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := p.emitter.EmitFromCandidate(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// Buffer returns the ring buffer for inspection.
func (p *Processor) Buffer() *ring.Buffer {
	return p.buffer
}

// Detector returns the spike detector.
func (p *Processor) Detector() *spike.Detector {
	return p.detector
}

// Gate returns the throttle gate.
func (p *Processor) Gate() *throttle.Gate {
	return p.gate
}

// Config returns a copy of the active configuration.
func (p *Processor) Config() config.Config {
	return p.cfg
}

// TickTimeouts finalizes spikes that exceeded the idle timeout.
func (p *Processor) TickTimeouts(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.clock.Now()
	for _, c := range p.detector.Tick(now) {
		if err := p.tryEmit(ctx, c); err != nil {
			return err
		}
	}
	return nil
}
