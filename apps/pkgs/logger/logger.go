package logger

import (
	"context"
	"log/slog"
	"os"
)

var log *slog.Logger

// Init initializes the JSON structured logger.
func Init() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)
}

// --- Leveled log functions ---

func Debug(msg string, args ...any) { log.Debug(msg, args...) }
func Info(msg string, args ...any)  { log.Info(msg, args...) }
func Warn(msg string, args ...any)  { log.Warn(msg, args...) }
func Error(msg string, args ...any) { log.Error(msg, args...) }

// --- Per-request log field accumulation (5W1H context) ---

// requestLogKey is the unexported context key for per-request log fields.
type requestLogKey struct{}

// RequestLog is a mutable struct that accumulates identity and auth context
// as a request flows through the middleware chain.
//
// It is stored as a pointer in context so that each middleware layer can
// write its fields and the HTTP Logger middleware can read them all after
// next.ServeHTTP() returns — without any mutex, because each request has
// its own pointer.
//
// 5W1H mapping:
//
//	WHO  → UserID, OrgID, AuthMethod
//	HOW  → RequestID (set by Logger for convenience of downstream loggers)
type RequestLog struct {
	// WHO — identity context populated by auth/RBAC middleware layers.
	UserID     string // authenticated user UUID ("" when unknown)
	OrgID      string // active organization UUID ("" when unknown)
	AuthMethod string // "bearer" | "apikey" | "bypass" | ""

	// HOW — set by the Logger middleware from chi's request ID.
	RequestID string
}

// NewRequestLog attaches a fresh *RequestLog to ctx and returns both.
// Call this at the outermost middleware layer (the HTTP Logger).
func NewRequestLog(ctx context.Context) (context.Context, *RequestLog) {
	rl := &RequestLog{}
	return context.WithValue(ctx, requestLogKey{}, rl), rl
}

// RequestLogFrom retrieves the *RequestLog from ctx.
// Returns a zero-value non-nil pointer (safe no-op) when not initialized.
func RequestLogFrom(ctx context.Context) *RequestLog {
	if rl, ok := ctx.Value(requestLogKey{}).(*RequestLog); ok {
		return rl
	}
	return &RequestLog{}
}
