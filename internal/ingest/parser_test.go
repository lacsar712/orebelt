package ingest

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lacsar712/orebelt/internal/model"
)

type stubProcessor struct {
	got []model.Sample
}

func (s *stubProcessor) Process(ctx context.Context, sample model.Sample) error {
	s.got = append(s.got, sample)
	return nil
}

func (s *stubProcessor) ProcessBatch(ctx context.Context, samples []model.Sample) error {
	s.got = append(s.got, samples...)
	return nil
}

func TestParseSample(t *testing.T) {
	s, err := ParseSample(SamplePayload{BeltID: " 3 ", Accel: 2.5, TS: "2026-01-01T00:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if s.BeltID != "3" || s.Accel != 2.5 {
		t.Fatalf("unexpected sample: %+v", s)
	}
}

func TestDecodeBatchSingle(t *testing.T) {
	body := bytes.NewBufferString(`{"belt_id":"1","accel":2.0}`)
	req := httptest.NewRequest("POST", "/v1/samples", body)
	samples, err := DecodeBatch(req)
	if err != nil || len(samples) != 1 {
		t.Fatalf("decode failed: %v len=%d", err, len(samples))
	}
}

func TestHandlerPostSamples(t *testing.T) {
	proc := &stubProcessor{}
	svc := NewService(proc)
	h := NewHandler(svc, nil)
	payload := `{"samples":[{"belt_id":"1","accel":3,"ts":"` + time.Now().Format(time.RFC3339) + `"}]}`
	req := httptest.NewRequest("POST", "/v1/samples", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()
	h.PostSamples(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(proc.got) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(proc.got))
	}
}

func TestValidateSample(t *testing.T) {
	err := ValidateSample(model.Sample{BeltID: "1", Accel: 1, TS: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
}
