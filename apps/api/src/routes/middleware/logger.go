package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"utils/logger"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// responseWriter wraps http.ResponseWriter to capture the status code written
// by downstream handlers, so the Logger can include it in the log entry.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger is an HTTP middleware that emits one structured log entry per request
// using a unified 5W1H (Who-What-When-Where-Why-How) schema.
//
// JSON output example:
//
//	{
//	  "time":  "2026-03-09T15:56:51.123Z",
//	  "level": "INFO",
//	  "msg":   "http.request",
//	  "who":   { "user_id": "uuid", "org_id": "uuid", "auth": "bearer", "ip": "10.0.1.5" },
//	  "what":  { "action": "GET /organizations/{org_slug}" },
//	  "when":  { "duration_ms": 42 },
//	  "where": { "layer": "http", "component": "Logger" },
//	  "why":   { "outcome": "success", "status": 200 },
//	  "how":   { "method": "GET", "path": "/api/v1/organizations/my-org",
//	             "route": "/api/v1/organizations/{org_slug}",
//	             "request_id": "abc-123", "user_agent": "...", "query": "" }
//	}
//
// WHO fields are populated by downstream auth/RBAC middleware via RequestLog.
// The chi RouteContext is pre-created here so chi fills in the route pattern
// in-place — making it readable after next.ServeHTTP() returns.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Pre-create chi RouteContext so chi populates RoutePattern in-place.
		// Chi reuses an existing RouteContext found in the request context.
		rctx := chi.NewRouteContext()
		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)

		// Attach mutable RequestLog for downstream middleware to write WHO/HOW fields.
		ctx, rl := logger.NewRequestLog(ctx)

		// Propagate the enriched request ID into RequestLog for downstream error logs.
		rl.RequestID = chimw.GetReqID(ctx)

		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)

		// — Collect 5W1H fields —

		duration := time.Since(start)
		status := rw.statusCode
		routePattern := rctx.RoutePattern()
		rawQuery := r.URL.RawQuery

		// Emit at the appropriate log level so 4xx/5xx responses surface in
		// alerting pipelines that filter by level rather than by outcome field.
		logFn := logLevelForStatus(status)
		logFn("http.request",
			// WHO — identity (populated by BearerAuth / ApiKeyAuth / RequireRole)
			slog.Group("who",
				slog.String("user_id", rl.UserID),
				slog.String("org_id", rl.OrgID),
				slog.String("auth", rl.AuthMethod),
				slog.String("ip", r.RemoteAddr),
			),
			// WHAT — the operation (route pattern for cardinality-safe aggregation)
			slog.Group("what",
				slog.String("action", fmt.Sprintf("%s %s", r.Method, stripAPIPrefix(routePattern))),
			),
			// WHEN — timing
			slog.Group("when",
				slog.Int64("duration_ms", duration.Milliseconds()),
			),
			// WHERE — system location
			slog.Group("where",
				slog.String("layer", "http"),
				slog.String("component", "Logger"),
			),
			// WHY — outcome (status-based classification for alerting)
			slog.Group("why",
				slog.String("outcome", outcomeFromStatus(status)),
				slog.Int("status", status),
			),
			// HOW — transport details
			slog.Group("how",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("route", routePattern),
				slog.String("request_id", rl.RequestID),
				slog.String("user_agent", r.UserAgent()),
				slog.String("query", rawQuery),
			),
		)
	})
}

// logFunc is the signature shared by logger.Info / Warn / Error.
type logFunc func(msg string, args ...any)

// logLevelForStatus returns the log function that matches the response severity:
//
//	2xx/3xx → Info   (normal traffic)
//	4xx     → Warn   (client mistake, may indicate abuse or misconfiguration)
//	5xx     → Error  (server fault, always requires attention)
func logLevelForStatus(status int) logFunc {
	switch {
	case status >= 500:
		return logger.Error
	case status >= 400:
		return logger.Warn
	default:
		return logger.Info
	}
}

// outcomeFromStatus maps an HTTP status code to a human-readable outcome label
// suitable for log-based metrics and alerts.
func outcomeFromStatus(status int) string {
	switch {
	case status < 400:
		return "success"
	case status < 500:
		return "client_error"
	default:
		return "server_error"
	}
}

// stripAPIPrefix removes the /api/v1 prefix from route patterns so the
// "what.action" field stays concise in dashboards and alert rules.
func stripAPIPrefix(pattern string) string {
	const prefix = "/api/v1"
	if strings.HasPrefix(pattern, prefix) {
		return pattern[len(prefix):]
	}
	return pattern
}
