package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestRateLimit(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	type args struct {
		cfg          RateLimiterConfig
		numRequests  int
		headers      map[string]string
		remoteAddr   string
		delayBetween time.Duration
	}
	type expected struct {
		finalStatusCode int
		hasRetryAfter   bool
		hasRateLimitHdr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "single request within limit succeeds",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 10},
				numRequests: 1,
				remoteAddr:  "192.168.1.1:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusOK,
				hasRetryAfter:   false,
				hasRateLimitHdr: true,
			},
		},
		{
			testName: "multiple requests within burst limit succeed",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 10},
				numRequests: 5,
				remoteAddr:  "192.168.1.2:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusOK,
				hasRetryAfter:   false,
				hasRateLimitHdr: true,
			},
		},

		// 異常系 (Error Cases)
		{
			testName: "burst of requests exceeds limit returns 429",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 3},
				numRequests: 5,
				remoteAddr:  "192.168.1.3:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusTooManyRequests,
				hasRetryAfter:   true,
				hasRateLimitHdr: true,
			},
		},

		// 境界値 (Boundary Values)
		{
			testName: "exactly at burst limit succeeds",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 5},
				numRequests: 5,
				remoteAddr:  "192.168.1.4:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusOK,
				hasRetryAfter:   false,
				hasRateLimitHdr: true,
			},
		},
		{
			testName: "one over burst limit fails",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 5},
				numRequests: 6,
				remoteAddr:  "192.168.1.5:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusTooManyRequests,
				hasRetryAfter:   true,
				hasRateLimitHdr: true,
			},
		},
		{
			testName: "burst size of 1 allows only one request",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 1},
				numRequests: 2,
				remoteAddr:  "192.168.1.6:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusTooManyRequests,
				hasRetryAfter:   true,
				hasRateLimitHdr: true,
			},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "API key with special characters identifies client",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 10},
				numRequests: 1,
				headers:     map[string]string{"X-API-Key": "key-with-特殊文字-🔑"},
				remoteAddr:  "192.168.1.7:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusOK,
				hasRetryAfter:   false,
				hasRateLimitHdr: true,
			},
		},

		// 空文字 (Empty/whitespace)
		{
			testName: "empty API key falls back to IP",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 10},
				numRequests: 1,
				headers:     map[string]string{"X-API-Key": ""},
				remoteAddr:  "192.168.1.8:12345",
			},
			expected: expected{
				finalStatusCode: http.StatusOK,
				hasRetryAfter:   false,
				hasRateLimitHdr: true,
			},
		},

		// Null/Nil (no headers at all)
		{
			testName: "no auth headers uses IP fallback",
			args: args{
				cfg:         RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 10},
				numRequests: 1,
				remoteAddr:  "10.0.0.1:54321",
			},
			expected: expected{
				finalStatusCode: http.StatusOK,
				hasRetryAfter:   false,
				hasRateLimitHdr: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mw := RateLimit(tt.args.cfg)
			handler := mw(okHandler)

			var lastResp *http.Response
			for i := 0; i < tt.args.numRequests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = tt.args.remoteAddr
				for k, v := range tt.args.headers {
					req.Header.Set(k, v)
				}
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				lastResp = w.Result()

				if tt.args.delayBetween > 0 && i < tt.args.numRequests-1 {
					time.Sleep(tt.args.delayBetween)
				}
			}

			if diff := cmp.Diff(tt.expected.finalStatusCode, lastResp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			retryAfter := lastResp.Header.Get("Retry-After")
			if tt.expected.hasRetryAfter && retryAfter == "" {
				t.Error("expected Retry-After header but not found")
			}
			if !tt.expected.hasRetryAfter && retryAfter != "" {
				t.Errorf("unexpected Retry-After header: %s", retryAfter)
			}

			if tt.expected.hasRateLimitHdr {
				if lastResp.Header.Get("X-RateLimit-Limit") == "" {
					t.Error("expected X-RateLimit-Limit header but not found")
				}
				if lastResp.Header.Get("X-RateLimit-Remaining") == "" {
					t.Error("expected X-RateLimit-Remaining header but not found")
				}
				if lastResp.Header.Get("X-RateLimit-Reset") == "" {
					t.Error("expected X-RateLimit-Reset header but not found")
				}
			}
		})
	}
}

func TestRateLimit_DifferentClientsIndependent(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 2}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	// Exhaust client A's burst
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Client B should still have full burst
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
		t.Errorf("client B should not be rate limited (-want +got):\n%s", diff)
	}
}

func TestRateLimit_BearerTokenClientIdentification(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 2}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	// Two requests from different IPs but same Bearer token share the limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0." + strconv.Itoa(i+1) + ":1234"
		req.Header.Set("Authorization", "Bearer shared-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Third request with same token should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.99:1234"
	req.Header.Set("Authorization", "Bearer shared-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if diff := cmp.Diff(http.StatusTooManyRequests, w.Result().StatusCode); diff != "" {
		t.Errorf("shared bearer token should be rate limited (-want +got):\n%s", diff)
	}
}

func TestRateLimit_APIKeyClientIdentification(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 2}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	// API key takes priority over Bearer token
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-API-Key", "my-api-key")
		req.Header.Set("Authorization", "Bearer some-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Third request with same API key should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-API-Key", "my-api-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if diff := cmp.Diff(http.StatusTooManyRequests, w.Result().StatusCode); diff != "" {
		t.Errorf("API key client should be rate limited (-want +got):\n%s", diff)
	}

	// Different API key should NOT be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	req2.Header.Set("X-API-Key", "different-api-key")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if diff := cmp.Diff(http.StatusOK, w2.Result().StatusCode); diff != "" {
		t.Errorf("different API key should not be rate limited (-want +got):\n%s", diff)
	}
}

func TestRateLimit_429ResponseBody(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 1}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	// First request uses the burst
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Second request gets 429
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	resp := w2.Result()
	if diff := cmp.Diff(http.StatusTooManyRequests, resp.StatusCode); diff != "" {
		t.Fatalf("expected 429 (-want +got):\n%s", diff)
	}

	contentType := resp.Header.Get("Content-Type")
	if diff := cmp.Diff("application/json", contentType); diff != "" {
		t.Errorf("content type mismatch (-want +got):\n%s", diff)
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Error == "" {
		t.Error("expected non-empty error message in response body")
	}
}

func TestRateLimit_RemainingHeaderDecreases(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 5}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	var remainings []int
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		remaining, err := strconv.Atoi(w.Result().Header.Get("X-RateLimit-Remaining"))
		if err != nil {
			t.Fatalf("failed to parse X-RateLimit-Remaining: %v", err)
		}
		remainings = append(remainings, remaining)
	}

	// Remaining should decrease with each request
	for i := 1; i < len(remainings); i++ {
		if remainings[i] >= remainings[i-1] {
			t.Errorf("remaining should decrease: got %v", remainings)
		}
	}
}

func TestRateLimit_TokenRefillAfterTime(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 600 requests per minute = 10 per second
	cfg := RateLimiterConfig{RequestsPerMinute: 600, BurstSize: 1}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	// Use the single token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
		t.Fatalf("first request should succeed (-want +got):\n%s", diff)
	}

	// Wait for token to refill (10/sec means ~100ms per token, wait a bit more)
	time.Sleep(200 * time.Millisecond)

	// Should succeed again after refill
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if diff := cmp.Diff(http.StatusOK, w2.Result().StatusCode); diff != "" {
		t.Errorf("request after token refill should succeed (-want +got):\n%s", diff)
	}
}

func TestRateLimit_ConcurrentRequestsSameClient(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := RateLimiterConfig{RequestsPerMinute: 60, BurstSize: 5}
	mw := RateLimit(cfg)
	handler := mw(okHandler)

	var wg sync.WaitGroup
	results := make([]int, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			results[idx] = w.Result().StatusCode
		}(i)
	}
	wg.Wait()

	okCount := 0
	rateLimitedCount := 0
	for _, code := range results {
		switch code {
		case http.StatusOK:
			okCount++
		case http.StatusTooManyRequests:
			rateLimitedCount++
		default:
			t.Errorf("unexpected status code: %d", code)
		}
	}

	// With burst of 5, at most 5 should succeed (possibly a few more due to refill)
	if okCount > 7 {
		t.Errorf("too many successful requests: got %d, expected at most ~7", okCount)
	}
	if rateLimitedCount == 0 {
		t.Error("expected some requests to be rate limited")
	}
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	cfg := DefaultRateLimiterConfig()

	if cfg.RequestsPerMinute <= 0 {
		t.Errorf("RequestsPerMinute should be positive, got %d", cfg.RequestsPerMinute)
	}
	if cfg.BurstSize <= 0 {
		t.Errorf("BurstSize should be positive, got %d", cfg.BurstSize)
	}
	if cfg.BurstSize > cfg.RequestsPerMinute {
		t.Errorf("BurstSize (%d) should not exceed RequestsPerMinute (%d)", cfg.BurstSize, cfg.RequestsPerMinute)
	}
}
