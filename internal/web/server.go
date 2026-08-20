package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/lacsar712/orebelt/internal/config"
	"github.com/lacsar712/orebelt/internal/emit"
	"github.com/lacsar712/orebelt/internal/ingest"
	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
	"github.com/lacsar712/orebelt/internal/ring"
	"github.com/lacsar712/orebelt/internal/throttle"
)

//go:embed static/*
var staticFS embed.FS

// Deps bundles dependencies required by the HTTP layer.
type Deps struct {
	Config   config.Config
	Ingest   *ingest.Handler
	Store    *emit.Store
	Buffer   *ring.Buffer
	Registry interface {
		List() []model.BeltMeta
	}
	Gate   *throttle.Gate
	Logger *logx.Logger
}

// Server serves REST endpoints and the embedded dashboard.
type Server struct {
	deps Deps
	mux  *http.ServeMux
}

// NewServer builds route handlers.
func NewServer(deps Deps) *Server {
	s := &Server{deps: deps, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/v1/samples", s.deps.Ingest.PostSamples)
	s.mux.HandleFunc("/v1/belts/", s.handleBelts)
	s.mux.HandleFunc("/v1/health", s.deps.Ingest.Health)
	s.mux.HandleFunc("/v1/stats", s.handleStats)
	sub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "dashboard unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleBelts(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/belts/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[1] != "spikes" {
		http.NotFound(w, r)
		return
	}
	beltID := parts[0]
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 {
			limit = n
		}
	}
	events := s.deps.Store.Recent(beltID, limit)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"belt_id": beltID,
		"spikes":  events,
		"count":   len(events),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	stats := map[string]any{
		"belts":   s.deps.Buffer.CollectStats(),
		"throttle": throttle.CollectStats(s.deps.Gate),
		"spikes":  s.deps.Store.AllSummaries(),
	}
	_ = json.NewEncoder(w).Encode(stats)
}
