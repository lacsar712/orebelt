package app

import (
	"encoding/json"
	"net/http"

	"github.com/lacsar712/orebelt/internal/ring"
	"github.com/lacsar712/orebelt/internal/spike"
)

// HealthHandler exposes detailed readiness information.
type HealthHandler struct {
	Buffer   *ring.Buffer
	Detector *spike.Detector
	Metrics  *Metrics
}

// ServeHTTP implements http.Handler.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"status": "ok",
		"belts":  len(h.Buffer.BeltIDs()),
	}
	if h.Metrics != nil {
		payload["metrics"] = h.Metrics.Snapshot()
	}
	if h.Detector != nil {
		payload["detector"] = map[string]any{
			"threshold_g": h.Detector.Params().ThresholdG,
			"min_samples": h.Detector.Params().MinSamples,
			"end_samples": h.Detector.Params().EndSamples,
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(payload)
}
