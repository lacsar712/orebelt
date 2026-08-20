package web

import (
	"net/http"
	"time"
)

// Middleware wraps an http.Handler with cross-cutting behavior.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestLogger logs method, path, and duration for each request.
func RequestLogger(logger interface {
	Info(format string, args ...any)
}) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			if logger != nil {
				logger.Info("%s %s %s", r.Method, r.URL.Path, time.Since(start))
			}
		})
	}
}

// RecoverPanic converts panics into HTTP 500 responses.
func RecoverPanic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds baseline response headers.
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			next.ServeHTTP(w, r)
		})
	}
}
