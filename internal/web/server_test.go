package web

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/lacsar712/orebelt/internal/config"
	"github.com/lacsar712/orebelt/internal/emit"
	"github.com/lacsar712/orebelt/internal/ingest"
	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/ring"
	"github.com/lacsar712/orebelt/internal/throttle"
)

type stubRegistry struct{}

func (stubRegistry) List() []model.BeltMeta { return nil }

type stubProcessor struct{}

func (stubProcessor) Process(ctx context.Context, s model.Sample) error { return nil }

func (stubProcessor) ProcessBatch(ctx context.Context, ss []model.Sample) error { return nil }

func TestIndexAndStats(t *testing.T) {
	cfg := config.Default()
	buf := ring.New(100)
	store := emit.NewStore(10)
	gate := throttle.NewGate(cfg.ThrottleInterval)
	svc := ingest.NewService(stubProcessor{})
	h := ingest.NewHandler(svc, logx.New())
	srv := NewServer(Deps{Config: cfg, Ingest: h, Store: store, Buffer: buf, Registry: stubRegistry{}, Gate: gate})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("index status=%d", rec.Code)
	}

	req2 := httptest.NewRequest("GET", "/v1/stats", nil)
	rec2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("stats status=%d", rec2.Code)
	}
}
