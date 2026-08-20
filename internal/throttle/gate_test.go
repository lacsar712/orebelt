package throttle

import (
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/spike"
)

func candidate(peak float64) spike.Candidate {
	now := time.Now()
	return spike.Candidate{BeltID: "3", Peak: peak, Start: now, End: now, SampleCnt: 3}
}

func TestGateAllowAndPending(t *testing.T) {
	g := NewGate(30 * time.Second)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c1 := candidate(3.0)
	if !g.Allow("3", now, c1) {
		t.Fatal("first emit should be allowed")
	}
	c2 := candidate(4.0)
	if g.Allow("3", now.Add(time.Second), c2) {
		t.Fatal("second emit should be throttled")
	}
	pending := g.Pending()
	if pending["3"].Peak != 4.0 {
		t.Fatalf("pending peak not refreshed: %+v", pending)
	}
}

func TestGateFlushDue(t *testing.T) {
	g := NewGate(10 * time.Second)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	g.Allow("3", start, candidate(3))
	g.Allow("3", start.Add(time.Second), candidate(5))
	due := g.FlushDue(start.Add(11 * time.Second))
	if len(due) != 1 || due[0].Peak != 5 {
		t.Fatalf("unexpected flush: %+v", due)
	}
}

func TestPendingPeakNotLowered(t *testing.T) {
	g := NewGate(time.Minute)
	now := time.Now()
	g.Allow("1", now, candidate(5))
	g.Allow("1", now.Add(time.Second), candidate(4))
	g.Allow("1", now.Add(2*time.Second), candidate(3))
	p := g.Pending()["1"]
	if p.Peak != 4 {
		t.Fatalf("peak should remain 4, got %.2f", p.Peak)
	}
}
