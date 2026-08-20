package ingest

import (
	"context"

	"github.com/lacsar712/orebelt/internal/model"
)

// Processor handles validated samples entering the pipeline.
type Processor interface {
	Process(ctx context.Context, s model.Sample) error
	ProcessBatch(ctx context.Context, samples []model.Sample) error
}

// Service validates and forwards samples to the application processor.
type Service struct {
	processor Processor
}

// NewService wires ingest to the downstream processor.
func NewService(p Processor) *Service {
	return &Service{processor: p}
}

// IngestOne accepts a single sample.
func (s *Service) IngestOne(ctx context.Context, sample model.Sample) error {
	if !sample.Valid() {
		return ErrInvalidSample
	}
	return s.processor.Process(ctx, sample)
}

// IngestMany accepts multiple samples in order.
func (s *Service) IngestMany(ctx context.Context, samples []model.Sample) error {
	if len(samples) == 0 {
		return ErrEmptyBatch
	}
	return s.processor.ProcessBatch(ctx, samples)
}

// ErrInvalidSample indicates malformed input.
var ErrInvalidSample = errString("invalid sample")

// ErrEmptyBatch indicates an empty batch request.
var ErrEmptyBatch = errString("empty batch")

type errString string

func (e errString) Error() string { return string(e) }
