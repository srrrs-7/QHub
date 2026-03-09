package cache

import (
	"context"
	"os"
	"testing"
	"time"
)

// --- Nil / disabled client tests (no Redis required) ---

func TestNew_EmptyURL(t *testing.T) {
	c := New("")
	if c != nil {
		t.Error("expected nil client for empty URL")
	}
}

func TestNew_InvalidURL(t *testing.T) {
	c := New("not-a-valid-url://???")
	if c != nil {
		t.Error("expected nil client for invalid URL")
	}
}

func TestNilClient_Available(t *testing.T) {
	var c *Client
	if c.Available() {
		t.Error("nil client should not be available")
	}
}

func TestNilClient_Get(t *testing.T) {
	var c *Client
	found, err := c.Get(context.Background(), "key", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected found=false for nil client")
	}
}

func TestNilClient_Set(t *testing.T) {
	var c *Client
	err := c.Set(context.Background(), "key", "value", time.Minute)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNilClient_Delete(t *testing.T) {
	var c *Client
	err := c.Delete(context.Background(), "key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNilClient_Ping(t *testing.T) {
	var c *Client
	err := c.Ping(context.Background())
	if err == nil {
		t.Error("expected error from Ping on nil client")
	}
}

func TestNilClient_Close(t *testing.T) {
	var c *Client
	err := c.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Integration tests (require running Redis) ---

func redisClient(t *testing.T) *Client {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://cache:6379"
	}
	c := New(url)
	if c == nil {
		// Try localhost as fallback
		c = New("redis://localhost:6379")
	}
	if c == nil {
		t.Skip("Redis not available")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		t.Skip("Redis not reachable:", err)
	}
	return c
}

func TestIntegration_SetGetRoundtrip(t *testing.T) {
	c := redisClient(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:roundtrip:" + t.Name()

	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	input := sample{Name: "hello", Count: 42}
	if err := c.Set(ctx, key, input, 10*time.Second); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	var got sample
	found, err := c.Get(ctx, key, &got)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if got.Name != input.Name || got.Count != input.Count {
		t.Errorf("got %+v, want %+v", got, input)
	}

	// Cleanup
	_ = c.Delete(ctx, key)
}

func TestIntegration_GetNonExistent(t *testing.T) {
	c := redisClient(t)
	defer c.Close()

	var dest string
	found, err := c.Get(context.Background(), "test:nonexistent:key:12345", &dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected found=false for non-existent key")
	}
}

func TestIntegration_Delete(t *testing.T) {
	c := redisClient(t)
	defer c.Close()

	ctx := context.Background()
	key := "test:delete:" + t.Name()

	_ = c.Set(ctx, key, "value", 10*time.Second)

	if err := c.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var dest string
	found, err := c.Get(ctx, key, &dest)
	if err != nil {
		t.Fatalf("Get after Delete failed: %v", err)
	}
	if found {
		t.Error("expected found=false after delete")
	}
}

func TestIntegration_Ping(t *testing.T) {
	c := redisClient(t)
	defer c.Close()

	if err := c.Ping(context.Background()); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}
