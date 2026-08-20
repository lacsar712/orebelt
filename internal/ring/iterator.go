package ring

import "github.com/lacsar712/orebelt/internal/model"

// Iterator walks samples in time order without exposing internal storage.
type Iterator struct {
	samples []model.Sample
	index   int
}

// NewIterator creates an iterator over a snapshot copy.
func NewIterator(samples []model.Sample) *Iterator {
	cp := make([]model.Sample, len(samples))
	copy(cp, samples)
	return &Iterator{samples: cp}
}

// Next returns the next sample and whether iteration should continue.
func (it *Iterator) Next() (model.Sample, bool) {
	if it.index >= len(it.samples) {
		return model.Sample{}, false
	}
	s := it.samples[it.index]
	it.index++
	return s, true
}

// Remaining reports how many samples are left.
func (it *Iterator) Remaining() int {
	if it.index >= len(it.samples) {
		return 0
	}
	return len(it.samples) - it.index
}

// Reset rewinds the iterator to the beginning.
func (it *Iterator) Reset() {
	it.index = 0
}

// Collect gathers all remaining samples into a new slice.
func (it *Iterator) Collect() []model.Sample {
	out := make([]model.Sample, 0, it.Remaining())
	for {
		s, ok := it.Next()
		if !ok {
			break
		}
		out = append(out, s)
	}
	it.Reset()
	return out
}

// Filter returns samples matching the predicate.
func Filter(samples []model.Sample, pred func(model.Sample) bool) []model.Sample {
	out := make([]model.Sample, 0)
	for _, s := range samples {
		if pred(s) {
			out = append(out, s)
		}
	}
	return out
}

// Latest returns the newest sample in a snapshot, if any.
func Latest(samples []model.Sample) (model.Sample, bool) {
	if len(samples) == 0 {
		return model.Sample{}, false
	}
	return samples[len(samples)-1], true
}
