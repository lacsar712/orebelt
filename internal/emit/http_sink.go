package emit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lacsar712/orebelt/internal/logx"
	"github.com/lacsar712/orebelt/internal/model"
)

// HTTPSink forwards spike events to an external maintenance endpoint.
type HTTPSink struct {
	client  *http.Client
	url     string
	logger  *logx.Logger
	retries int
}

// NewHTTPSink creates an HTTP emitter. Empty url disables remote forwarding.
func NewHTTPSink(url string, logger *logx.Logger) *HTTPSink {
	return &HTTPSink{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url:     url,
		logger:  logger,
		retries: 2,
	}
}

// Enabled reports whether remote forwarding is configured.
func (h *HTTPSink) Enabled() bool {
	return h.url != ""
}

// Send posts one spike event, honoring context cancellation.
func (h *HTTPSink) Send(ctx context.Context, ev model.SpikeEvent) error {
	if !h.Enabled() {
		return nil
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal spike event: %w", err)
	}
	var lastErr error
	for attempt := 0; attempt <= h.retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, h.url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			if h.logger != nil {
				h.logger.Warn("emit attempt %d failed: %v", attempt+1, err)
			}
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return lastErr
}
