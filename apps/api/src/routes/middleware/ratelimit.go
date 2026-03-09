package middleware

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiterConfig holds rate limiting configuration.
type RateLimiterConfig struct {
	// RequestsPerMinute is the sustained rate of allowed requests per client.
	RequestsPerMinute int
	// BurstSize is the maximum number of requests a client can make in a burst.
	BurstSize int
}

// DefaultRateLimiterConfig returns sensible defaults for rate limiting.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	}
}

// tokenBucket tracks rate limit state for a single client.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	lastAccess time.Time
}

// RateLimit returns middleware that limits requests per client using a token bucket algorithm.
// Client is identified by: X-API-Key header > Bearer token > RemoteAddr (fallback).
// Returns 429 Too Many Requests with Retry-After header when the limit is exceeded.
func RateLimit(cfg RateLimiterConfig) func(http.Handler) http.Handler {
	var clients sync.Map
	refillRate := float64(cfg.RequestsPerMinute) / 60.0 // tokens per second

	// Background cleanup goroutine to remove stale entries
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			staleThreshold := time.Now().Add(-10 * time.Minute)
			clients.Range(func(key, value any) bool {
				bucket := value.(*tokenBucket)
				bucket.mu.Lock()
				lastAccess := bucket.lastAccess
				bucket.mu.Unlock()
				if lastAccess.Before(staleThreshold) {
					clients.Delete(key)
				}
				return true
			})
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := extractClientKey(r)

			val, _ := clients.LoadOrStore(clientKey, &tokenBucket{
				tokens:     float64(cfg.BurstSize),
				lastRefill: time.Now(),
				lastAccess: time.Now(),
			})
			bucket := val.(*tokenBucket)

			bucket.mu.Lock()
			now := time.Now()
			bucket.lastAccess = now

			// Refill tokens based on elapsed time
			elapsed := now.Sub(bucket.lastRefill).Seconds()
			bucket.tokens += elapsed * refillRate
			if bucket.tokens > float64(cfg.BurstSize) {
				bucket.tokens = float64(cfg.BurstSize)
			}
			bucket.lastRefill = now

			// Calculate reset time (when bucket would be full again)
			tokensNeeded := float64(cfg.BurstSize) - bucket.tokens
			var resetSeconds int
			if tokensNeeded > 0 && refillRate > 0 {
				resetSeconds = int(math.Ceil(tokensNeeded / refillRate))
			}

			if bucket.tokens < 1 {
				// Not enough tokens
				retryAfter := 1.0
				if refillRate > 0 {
					retryAfter = math.Ceil((1.0 - bucket.tokens) / refillRate)
				}
				bucket.mu.Unlock()

				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.BurstSize))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.Itoa(resetSeconds))
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter)))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(errorResponse{Error: "rate limit exceeded"})
				return
			}

			// Consume a token
			bucket.tokens--
			remaining := int(bucket.tokens)
			if remaining < 0 {
				remaining = 0
			}
			bucket.mu.Unlock()

			// Set rate limit headers on successful requests
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.BurstSize))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(resetSeconds))

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientKey determines the client identifier from the request.
// Priority: X-API-Key header > Bearer token > RemoteAddr.
func extractClientKey(r *http.Request) string {
	// Priority 1: X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "apikey:" + apiKey
	}

	// Priority 2: Bearer token
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != "" {
			return "bearer:" + parts[1]
		}
	}

	// Priority 3: RemoteAddr (strip port)
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return "ip:" + addr
}
