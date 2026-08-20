package ingest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lacsar712/orebelt/internal/logx"
)

// Handler exposes HTTP endpoints for sample ingestion.
type Handler struct {
	service *Service
	logger  *logx.Logger
}

// NewHandler creates the ingest HTTP adapter.
func NewHandler(service *Service, logger *logx.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// PostSamples handles POST /v1/samples.
func (h *Handler) PostSamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	samples, err := DecodeBatch(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.service.IngestMany(r.Context(), samples); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidSample) || errors.Is(err, ErrEmptyBatch) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accepted": len(samples),
		"status":   "ok",
	})
}

// Health reports ingest readiness.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
