package model

import "strings"

// NormalizeBeltID trims whitespace and rejects empty belt identifiers.
func NormalizeBeltID(id string) (string, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", false
	}
	return id, true
}

// BeltMeta holds static metadata for a monitored belt segment.
type BeltMeta struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Location    string  `json:"location"`
	MaxSafeG    float64 `json:"max_safe_g"`
	SampleRateHz int    `json:"sample_rate_hz"`
}

// DefaultMeta returns a placeholder metadata record for unknown belts.
func DefaultMeta(id string) BeltMeta {
	return BeltMeta{
		ID:           id,
		Label:        "Belt " + id,
		Location:     "unknown",
		MaxSafeG:     5.0,
		SampleRateHz: 100,
	}
}
