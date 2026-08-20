package spike

import (
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

func sample(belt string, accel float64, sec int) model.Sample {
	return model.Sample{
		BeltID: belt,
		Accel:  accel,
		TS:     time.Date(2026, 3, 1, 0, 0, sec, 0, time.UTC),
	}
}

func TestFSMEnterAndExit(t *testing.T) {
	f := NewFSM(Params{ThresholdG: 2.5, MinSamples: 3, EndSamples: 2, Timeout: time.Second})
	var c *Candidate
	c = f.Observe(sample("1", 3, 0))
	if c != nil || f.State() != StateIdle {
		t.Fatalf("unexpected after hit 1: %v state=%s", c, f.State())
	}
	f.Observe(sample("1", 3.1, 1))
	f.Observe(sample("1", 3.2, 2))
	if f.State() != StateInSpike {
		t.Fatalf("expected in_spike, got %s", f.State())
	}
	f.Observe(sample("1", 1, 3))
	c = f.Observe(sample("1", 1, 4))
	if c == nil {
		t.Fatal("expected candidate on exit")
	}
	if c.Peak < 3.2 || c.SampleCnt < 3 {
		t.Fatalf("bad candidate: %+v", c)
	}
	if f.State() != StateIdle {
		t.Fatal("expected idle after exit")
	}
}

func TestFSMTimeout(t *testing.T) {
	f := NewFSM(Params{ThresholdG: 2.5, MinSamples: 2, EndSamples: 5, Timeout: 100 * time.Millisecond})
	f.Observe(sample("1", 3, 0))
	f.Observe(sample("1", 3, 1))
	last := sample("1", 3, 2)
	f.Observe(last)
	c := f.Tick(last.TS.Add(200 * time.Millisecond))
	if c == nil {
		t.Fatal("expected timeout candidate")
	}
}

func TestDetectorMultiBelt(t *testing.T) {
	d := NewDetector(Params{ThresholdG: 2.5, MinSamples: 2, EndSamples: 1})
	d.Observe(sample("a", 3, 0))
	c := d.Observe(sample("a", 3, 1))
	if c != nil {
		t.Fatal("spike not finished yet")
	}
	if d.BeltState("a") != StateInSpike {
		t.Fatal("belt a should be in spike")
	}
	if d.BeltState("b") != StateIdle {
		t.Fatal("belt b should be idle")
	}
}
