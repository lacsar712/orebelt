package app

import (
	"sort"
	"sync"

	"github.com/lacsar712/orebelt/internal/model"
)

// Registry tracks belt metadata known to the service.
type Registry struct {
	mu    sync.RWMutex
	belts map[string]model.BeltMeta
}

// NewRegistry creates an empty belt registry.
func NewRegistry() *Registry {
	return &Registry{belts: make(map[string]model.BeltMeta)}
}

// Register adds or replaces belt metadata.
func (r *Registry) Register(meta model.BeltMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.belts[meta.ID] = meta
}

// Get returns metadata for a belt, synthesizing defaults when unknown.
func (r *Registry) Get(id string) model.BeltMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if meta, ok := r.belts[id]; ok {
		return meta
	}
	return model.DefaultMeta(id)
}

// List returns all registered belt metadata sorted by ID.
func (r *Registry) List() []model.BeltMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.BeltMeta, 0, len(r.belts))
	for _, m := range r.belts {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Ensure registers a default entry when the belt is unseen.
func (r *Registry) Ensure(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.belts[id]; !ok {
		r.belts[id] = model.DefaultMeta(id)
	}
}
