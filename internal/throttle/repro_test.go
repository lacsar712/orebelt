package throttle

import (
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/spike"
)

// TestStorePendingKeepsHigherPeak reproduces orebelt-004.
func TestStorePendingKeepsHigherPeak(t *testing.T) {
	g := NewGate(time.Minute)
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if !g.Allow("belt-t", now, spike.Candidate{BeltID: "belt-t", Peak: 5.0, SampleCnt: 3}) {
		t.Fatal("first allow should succeed")
	}
	if g.Allow("belt-t", now.Add(time.Second), spike.Candidate{BeltID: "belt-t", Peak: 4.0, SampleCnt: 3}) {
		t.Fatal("second allow should throttle")
	}
	if g.Allow("belt-t", now.Add(2*time.Second), spike.Candidate{BeltID: "belt-t", Peak: 3.0, SampleCnt: 3}) {
		t.Fatal("third allow should throttle")
	}
	p := g.Pending()["belt-t"]
	if p.Peak != 4.0 {
		t.Fatalf("pending must keep higher peak (not overwrite with lower): got %.1f want 4.0", p.Peak)
	}

	st := CollectStats(g)
	if st.Pending["belt-t"].Peak != 4.0 {
		t.Fatalf("CollectStats pending peak want 4.0, got %.1f", st.Pending["belt-t"].Peak)
	}
}
