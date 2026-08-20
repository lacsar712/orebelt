package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lacsar712/orebelt/internal/config"
	"github.com/lacsar712/orebelt/internal/emit"
	"github.com/lacsar712/orebelt/internal/ingest"
	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/spike"
	"github.com/lacsar712/orebelt/internal/throttle"
	"github.com/lacsar712/orebelt/internal/timeutil"
	"github.com/lacsar712/orebelt/internal/web"
)

// App wires together configuration, processing, HTTP serving, and background tasks.
type App struct {
	cfg        config.Config
	logger     *logx.Logger
	clock      timeutil.Clock
	registry   *Registry
	store      *emit.Store
	emitter    *emit.Emitter
	processor  *Processor
	ingest     *ingest.Service
	scheduler  *throttle.Scheduler
	web        *web.Server
	httpServer *http.Server
	cancelBg   context.CancelFunc
	stopOnce   sync.Once
}

// New constructs an application from configuration.
func New(cfg config.Config, logger *logx.Logger) (*App, error) {
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = logx.New()
	}
	clock := timeutil.RealClock{}
	registry := NewRegistry()
	store := emit.NewStore(cfg.MaxRecentSpikes)
	sink := emit.NewHTTPSink(cfg.EmitURL, logger)
	emitter := emit.NewEmitter(store, sink, clock, logger)
	processor := NewProcessor(cfg, emitter, registry, clock, logger)
	ingestSvc := ingest.NewService(processor)

	a := &App{
		cfg:       cfg,
		logger:    logger,
		clock:     clock,
		registry:  registry,
		store:     store,
		emitter:   emitter,
		processor: processor,
		ingest:    ingestSvc,
	}

	a.scheduler = throttle.NewScheduler(processor.Gate(), cfg.FlushInterval, func(candidates []spike.Candidate) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := emitter.EmitBatch(ctx, candidates); err != nil && a.logger != nil {
			a.logger.Error("flush emit failed: %v", err)
		}
	})

	a.web = web.NewServer(web.Deps{
		Config:    cfg,
		Ingest:    ingest.NewHandler(ingestSvc, logger),
		Store:     store,
		Buffer:    processor.Buffer(),
		Registry:  registry,
		Gate:      processor.Gate(),
		Logger:    logger,
	})

	return a, nil
}

// Run starts background workers and the HTTP server until context cancellation.
func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	a.cancelBg = cancel

	a.scheduler.Start(a.clock.Now)
	go a.runTimeoutLoop(ctx)

	a.httpServer = &http.Server{
		Addr:         a.cfg.ListenAddr,
		Handler:      a.web.Handler(),
		ReadTimeout:  a.cfg.ReadTimeout,
		WriteTimeout: a.cfg.WriteTimeout,
		IdleTimeout:  a.cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("orebelt listening on %s", a.cfg.ListenAddr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return a.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (a *App) runTimeoutLoop(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.SpikeTimeout / 2)
	if ticker.C == nil {
		ticker = time.NewTicker(250 * time.Millisecond)
	}
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.processor.TickTimeouts(ctx); err != nil && a.logger != nil {
				a.logger.Warn("timeout tick failed: %v", err)
			}
		}
	}
}

// Shutdown gracefully stops background tasks and the HTTP server.
func (a *App) Shutdown(ctx context.Context) error {
	var err error
	a.stopOnce.Do(func() {
		if a.cancelBg != nil {
			a.cancelBg()
		}
		if a.scheduler != nil {
			a.scheduler.Stop()
		}
		if a.httpServer != nil {
			if shutdownErr := a.httpServer.Shutdown(ctx); shutdownErr != nil {
				err = shutdownErr
			}
		}
		if emitErr := a.emitter.Shutdown(ctx); emitErr != nil && err == nil {
			err = emitErr
		}
	})
	return err
}

// Handler exposes the HTTP handler for tests.
func (a *App) Handler() http.Handler {
	return a.web.Handler()
}

// Processor returns the sample processor for integration tests.
func (a *App) Processor() *Processor {
	return a.processor
}

// Store returns the spike event store.
func (a *App) Store() *emit.Store {
	return a.store
}

// Config returns active configuration.
func (a *App) Config() config.Config {
	return a.cfg
}

// String returns a short diagnostic label.
func (a *App) String() string {
	return fmt.Sprintf("orebelt@%s", a.cfg.ListenAddr)
}
