package spike

import (
	"sort"

	"github.com/lacsar712/orebelt/internal/model"
)

// Analyzer computes descriptive statistics over a sample window.
type Analyzer struct {
	ThresholdG float64
}

// Summary holds aggregate metrics for a belt window.
type Summary struct {
	BeltID     string  `json:"belt_id"`
	Count      int     `json:"count"`
	MeanAbs    float64 `json:"mean_abs"`
	MaxAbs     float64 `json:"max_abs"`
	HitsAbove  int     `json:"hits_above_threshold"`
	HitRatio   float64 `json:"hit_ratio"`
}

// Analyze returns statistics for the provided samples.
func (a Analyzer) Analyze(beltID string, samples []model.Sample) Summary {
	sum := Summary{BeltID: beltID, Count: len(samples)}
	if len(samples) == 0 {
		return sum
	}
	var total float64
	for _, s := range samples {
		abs := s.AbsAccel()
		total += abs
		if abs > sum.MaxAbs {
			sum.MaxAbs = abs
		}
		if abs > a.ThresholdG {
			sum.HitsAbove++
		}
	}
	sum.MeanAbs = total / float64(len(samples))
	sum.HitRatio = float64(sum.HitsAbove) / float64(len(samples))
	return sum
}

// RankPeaks returns samples sorted by absolute acceleration descending.
func RankPeaks(samples []model.Sample, limit int) []model.Sample {
	if limit <= 0 || len(samples) == 0 {
		return nil
	}
	cp := make([]model.Sample, len(samples))
	copy(cp, samples)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].AbsAccel() > cp[j].AbsAccel()
	})
	if limit > len(cp) {
		limit = len(cp)
	}
	out := make([]model.Sample, limit)
	copy(out, cp[:limit])
	return out
}

// Intervals finds contiguous above-threshold intervals in time order.
func Intervals(samples []model.Sample, threshold float64, minLen int) [][2]int {
	if minLen < 1 {
		minLen = 1
	}
	var out [][2]int
	start := -1
	for i, s := range samples {
		hit := s.AbsAccel() > threshold
		if hit && start < 0 {
			start = i
		}
		if !hit && start >= 0 {
			if i-start >= minLen {
				out = append(out, [2]int{start, i - 1})
			}
			start = -1
		}
	}
	if start >= 0 && len(samples)-start >= minLen {
		out = append(out, [2]int{start, len(samples) - 1})
	}
	return out
}
