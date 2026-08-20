package web

import (
	"encoding/json"
	"net/http"
)

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError returns a JSON error payload.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseLimit extracts a positive limit query parameter with fallback.
func parseLimit(r *http.Request, fallback int) int {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return fallback
	}
	var n int
	if _, err := json.Number(q).Int64(); err == nil {
		if v, _ := json.Number(q).Int64(); v > 0 {
			n = int(v)
		}
	}
	if n <= 0 {
		return fallback
	}
	return n
}
