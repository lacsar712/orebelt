package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

const maxBodyBytes = 1 << 20

// SamplePayload is the JSON body for POST /v1/samples.
type SamplePayload struct {
	BeltID string  `json:"belt_id"`
	Accel  float64 `json:"accel"`
	TS     string  `json:"ts"`
}

// BatchPayload accepts multiple samples in one request.
type BatchPayload struct {
	Samples []SamplePayload `json:"samples"`
}

// ParseSample converts a payload into a validated model.Sample.
func ParseSample(p SamplePayload) (model.Sample, error) {
	id, ok := model.NormalizeBeltID(p.BeltID)
	if !ok {
		return model.Sample{}, fmt.Errorf("belt_id is required")
	}
	ts, err := parseTimestamp(p.TS)
	if err != nil {
		return model.Sample{}, err
	}
	return model.Sample{BeltID: id, Accel: p.Accel, TS: ts}, nil
}

func parseTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC(), nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), nil
		}
	}
	if unix, err := parseUnix(raw); err == nil {
		return unix, nil
	}
	return time.Time{}, fmt.Errorf("invalid ts: %q", raw)
}

func parseUnix(raw string) (time.Time, error) {
	var sec int64
	var nsec int64
	if _, err := fmt.Sscanf(raw, "%d", &sec); err != nil {
		return time.Time{}, err
	}
	if sec > 1_000_000_000_000 {
		nsec = (sec % 1000) * int64(time.Millisecond)
		sec = sec / 1000
	}
	return time.Unix(sec, nsec).UTC(), nil
}

// DecodeSample reads one sample from an HTTP request body.
func DecodeSample(r *http.Request) (model.Sample, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return model.Sample{}, err
	}
	var p SamplePayload
	if err := json.Unmarshal(data, &p); err != nil {
		return model.Sample{}, fmt.Errorf("invalid json: %w", err)
	}
	return ParseSample(p)
}

// DecodeBatch reads a batch payload from an HTTP request body.
func DecodeBatch(r *http.Request) ([]model.Sample, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var batch BatchPayload
	if err := json.Unmarshal(data, &batch); err == nil && len(batch.Samples) > 0 {
		out := make([]model.Sample, 0, len(batch.Samples))
		for i, p := range batch.Samples {
			s, err := ParseSample(p)
			if err != nil {
				return nil, fmt.Errorf("samples[%d]: %w", i, err)
			}
			out = append(out, s)
		}
		return out, nil
	}
	var single SamplePayload
	if err := json.Unmarshal(data, &single); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	one, err := ParseSample(single)
	if err != nil {
		return nil, err
	}
	return []model.Sample{one}, nil
}
