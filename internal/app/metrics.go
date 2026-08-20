package app

import (
	"context"
	"sync/atomic"

	"github.com/lacsar712/orebelt/internal/model"
)

// Metrics tracks coarse runtime counters for observability.
type Metrics struct {
	samplesIngested atomic.Uint64
	spikesDetected  atomic.Uint64
	spikesEmitted   atomic.Uint64
	spikesThrottled atomic.Uint64
}

// NewMetrics creates zeroed metrics.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncSamples records accepted samples.
func (m *Metrics) IncSamples(n uint64) {
	m.samplesIngested.Add(n)
}

// IncDetected records spike candidates.
func (m *Metrics) IncDetected() {
	m.spikesDetected.Add(1)
}

// IncEmitted records successful emissions.
func (m *Metrics) IncEmitted() {
	m.spikesEmitted.Add(1)
}

// IncThrottled records throttled candidates.
func (m *Metrics) IncThrottled() {
	m.spikesThrottled.Add(1)
}

// Snapshot returns a copy of current counters.
func (m *Metrics) Snapshot() map[string]uint64 {
	return map[string]uint64{
		"samples_ingested": m.samplesIngested.Load(),
		"spikes_detected":  m.spikesDetected.Load(),
		"spikes_emitted":   m.spikesEmitted.Load(),
		"spikes_throttled": m.spikesThrottled.Load(),
	}
}

// InstrumentedProcessor wraps Processor with metrics collection.
type InstrumentedProcessor struct {
	*Processor
	metrics *Metrics
}

// NewInstrumentedProcessor decorates a processor with metrics.
func NewInstrumentedProcessor(p *Processor, m *Metrics) *InstrumentedProcessor {
	return &InstrumentedProcessor{Processor: p, metrics: m}
}

// ProcessBatch increments sample counters after successful processing.
func (ip *InstrumentedProcessor) ProcessBatch(ctx context.Context, samples []model.Sample) error {
	err := ip.Processor.ProcessBatch(ctx, samples)
	if err == nil {
		ip.metrics.IncSamples(uint64(len(samples)))
	}
	return err
}

// Process delegates to ProcessBatch.
func (ip *InstrumentedProcessor) Process(ctx context.Context, s model.Sample) error {
	return ip.ProcessBatch(ctx, []model.Sample{s})
}
