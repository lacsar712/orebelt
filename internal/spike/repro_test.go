package spike

import (
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

// TestFSMNegativeAccelUsesAbs reproduces orebelt-002.
func TestFSMNegativeAccelUsesAbs(t *testing.T) {
	if (model.Sample{Accel: -3.2}).AbsAccel() != 3.2 {
		t.Fatalf("AbsAccel(-3.2) want 3.2, got %v", (model.Sample{Accel: -3.2}).AbsAccel())
	}
	if (model.Sample{Accel: 3.2}).AbsAccel() != 3.2 {
		t.Fatalf("AbsAccel(3.2) want 3.2, got %v", (model.Sample{Accel: 3.2}).AbsAccel())
	}

	f := NewFSM(Params{ThresholdG: 2.5, MinSamples: 2, EndSamples: 1, Timeout: time.Second})
	f.Observe(sample("n", -3.0, 0))
	f.Observe(sample("n", -3.1, 1))
	if f.State() != StateInSpike {
		t.Fatalf("negative accel |g| above threshold must enter spike, state=%s", f.State())
	}

	a := Analyzer{ThresholdG: 2.5}
	sum := a.Analyze("n", []model.Sample{sample("n", -3.0, 0), sample("n", -1.0, 1)})
	if sum.HitsAbove != 1 || sum.MaxAbs != 3.0 {
		t.Fatalf("analyzer must count abs hits: %+v", sum)
	}
}

// TestFSMMinSamplesInclusive reproduces orebelt-003.
func TestFSMMinSamplesInclusive(t *testing.T) {
	f := NewFSM(Params{ThresholdG: 2.5, MinSamples: 3, EndSamples: 2, Timeout: time.Second})
	f.Observe(sample("m", 3.0, 0))
	f.Observe(sample("m", 3.0, 1))
	f.Observe(sample("m", 3.0, 2))
	if f.State() != StateInSpike {
		t.Fatalf("exactly MinSamples hits must enter spike (>=), state=%s", f.State())
	}

	d := NewDetector(Params{ThresholdG: 2.5, MinSamples: 3, EndSamples: 2, Timeout: time.Second})
	d.Observe(sample("d", 3.0, 0))
	d.Observe(sample("d", 3.0, 1))
	d.Observe(sample("d", 3.0, 2))
	if d.BeltState("d") != StateInSpike {
		t.Fatalf("detector must enter spike at exactly MinSamples, state=%s", d.BeltState("d"))
	}

	iv := Intervals([]model.Sample{
		sample("i", 3, 0), sample("i", 3, 1), sample("i", 3, 2), sample("i", 0, 3),
	}, 2.5, 3)
	if len(iv) != 1 || iv[0][0] != 0 || iv[0][1] != 2 {
		t.Fatalf("Intervals minLen must be inclusive (>=): %+v", iv)
	}
}

// TestFSMMissCounterResetsOnHit reproduces orebelt-006.
func TestFSMMissCounterResetsOnHit(t *testing.T) {
	f := NewFSM(Params{ThresholdG: 2.5, MinSamples: 2, EndSamples: 3, Timeout: time.Second})
	f.Observe(sample("x", 3.0, 0))
	f.Observe(sample("x", 3.0, 1))
	if f.State() != StateInSpike {
		t.Fatalf("expected in_spike, got %s", f.State())
	}
	// Two misses, then another hit — miss streak must reset.
	f.Observe(sample("x", 0.5, 2))
	f.Observe(sample("x", 0.5, 3))
	f.Observe(sample("x", 3.0, 4))
	if f.State() != StateInSpike {
		t.Fatalf("hit inside spike must reset consecutiveMiss, state=%s", f.State())
	}
	f.Observe(sample("x", 0.5, 5))
	f.Observe(sample("x", 0.5, 6))
	c := f.Observe(sample("x", 0.5, 7))
	if c == nil {
		t.Fatal("expected candidate after EndSamples consecutive misses following a hit reset")
	}
	if c.SampleCnt < 3 {
		t.Fatalf("candidate should include post-reset hit samples: %+v", c)
	}
}

// TestDetectorObserveSetsBeltID reproduces orebelt-009.
func TestDetectorObserveSetsBeltID(t *testing.T) {
	d := NewDetector(Params{ThresholdG: 2.5, MinSamples: 2, EndSamples: 1, Timeout: time.Second})
	d.Observe(sample("belt-9", 3.0, 0))
	d.Observe(sample("belt-9", 3.1, 1))
	c := d.Observe(sample("belt-9", 0.1, 2))
	if c == nil {
		t.Fatal("expected completed spike candidate")
	}
	if c.BeltID != "belt-9" {
		t.Fatalf("Observe must set candidate BeltID from sample, got %q", c.BeltID)
	}
}
