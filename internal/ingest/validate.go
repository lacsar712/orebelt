package ingest

import (
	"fmt"
	"math"

	"github.com/lacsar712/orebelt/internal/model"
)

// ValidateSample performs field-level validation beyond Valid().
func ValidateSample(s model.Sample) error {
	if !s.Valid() {
		return ErrInvalidSample
	}
	if math.IsNaN(s.Accel) || math.IsInf(s.Accel, 0) {
		return fmt.Errorf("accel must be finite")
	}
	if s.Accel < -1000 || s.Accel > 1000 {
		return fmt.Errorf("accel out of range")
	}
	return nil
}

// ValidateBatch validates every sample in a batch.
func ValidateBatch(samples []model.Sample) error {
	for i, s := range samples {
		if err := ValidateSample(s); err != nil {
			return fmt.Errorf("samples[%d]: %w", i, err)
		}
	}
	return nil
}
