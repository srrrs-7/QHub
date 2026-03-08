package parallel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewKeyShardWorker(t *testing.T) {
	type args struct {
		workerCnt int
		chBuf     int
	}
	type expected struct {
		isErr bool
	}
	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		{
			testName: "valid parameters",
			args:     args{workerCnt: 4, chBuf: 10},
			expected: expected{isErr: false},
		},
		{
			testName: "workerCnt is zero",
			args:     args{workerCnt: 0, chBuf: 10},
			expected: expected{isErr: true},
		},
		{
			testName: "workerCnt is negative",
			args:     args{workerCnt: -1, chBuf: 10},
			expected: expected{isErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ks, err := NewKeyShardWorker(tt.args.workerCnt, tt.args.chBuf)
			if tt.expected.isErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if ks != nil {
					t.Error("expected nil KeyShard on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if ks == nil {
					t.Error("expected non-nil KeyShard")
				}
			}
		})
	}
}

func TestKeyShardWorker_BasicFlow(t *testing.T) {
	ks, err := NewKeyShardWorker(4, 10)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	var processed sync.Map
	ctx := context.Background()

	ks.Work(ctx, func(_ context.Context, key string) error {
		processed.Store(key, true)
		return nil
	})

	keys := []string{"a", "b", "c", "d", "e"}
	for _, key := range keys {
		if err := ks.Send(ctx, key); err != nil {
			t.Fatalf("Send(%q): %v", key, err)
		}
	}

	if err := ks.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	for _, key := range keys {
		if _, ok := processed.Load(key); !ok {
			t.Errorf("key %q was not processed", key)
		}
	}
}

func TestKeyShardWorker_SameKeyOrderPreserved(t *testing.T) {
	ks, err := NewKeyShardWorker(4, 100)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	var mu sync.Mutex
	received := make(map[string][]string)
	ctx := context.Background()

	ks.Work(ctx, func(_ context.Context, key string) error {
		mu.Lock()
		received[key] = append(received[key], key)
		mu.Unlock()
		return nil
	})

	n := 20
	for range n {
		ks.Send(ctx, "alpha")
		ks.Send(ctx, "beta")
	}

	ks.Close()

	for _, key := range []string{"alpha", "beta"} {
		if len(received[key]) != n {
			t.Errorf("key %q: expected %d processed, got %d", key, n, len(received[key]))
		}
	}
}

func TestKeyShardWorker_SendAfterClose(t *testing.T) {
	ks, err := NewKeyShardWorker(2, 10)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	ctx := context.Background()
	ks.Work(ctx, func(_ context.Context, key string) error { return nil })
	ks.Close()

	err = ks.Send(ctx, "key")
	if !errors.Is(err, ErrClosed) {
		t.Errorf("expected ErrClosed, got %v", err)
	}
}

func TestKeyShardWorker_ErrorCollection(t *testing.T) {
	ks, err := NewKeyShardWorker(1, 10)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	errBoom := errors.New("boom")
	ctx := context.Background()

	ks.Work(ctx, func(_ context.Context, key string) error {
		if key == "fail" {
			return errBoom
		}
		return nil
	})

	ks.Send(ctx, "ok")
	ks.Send(ctx, "fail")

	closeErr := ks.Close()
	if closeErr == nil {
		t.Fatal("expected error from Close, got nil")
	}
	if !errors.Is(closeErr, errBoom) {
		t.Errorf("expected wrapped errBoom, got %v", closeErr)
	}
}

func TestKeyShardWorker_ContextCancelSend(t *testing.T) {
	ks, err := NewKeyShardWorker(1, 0)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = ks.Send(ctx, "key")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	ks.Work(context.Background(), func(_ context.Context, key string) error { return nil })
	ks.Close()
}

func TestKeyShardWorker_ContextCancelWork(t *testing.T) {
	ks, err := NewKeyShardWorker(2, 10)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	ks.Work(ctx, func(_ context.Context, key string) error { return nil })

	bgCtx := context.Background()
	ks.Send(bgCtx, "a")
	ks.Send(bgCtx, "b")

	time.Sleep(50 * time.Millisecond)
	cancel()

	done := make(chan error, 1)
	go func() { done <- ks.Close() }()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not return after context cancellation")
	}
}

func TestKeyShardWorker_WorkIdempotency(t *testing.T) {
	ks, err := NewKeyShardWorker(2, 10)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	var callCount atomic.Int32
	ctx := context.Background()

	fn := func(_ context.Context, key string) error {
		callCount.Add(1)
		return nil
	}

	ks.Work(ctx, fn)
	ks.Work(ctx, fn)

	ks.Send(ctx, "a")
	ks.Close()

	if count := callCount.Load(); count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}
}

func TestKeyShardWorker_CloseIdempotency(t *testing.T) {
	ks, err := NewKeyShardWorker(2, 10)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	ctx := context.Background()
	ks.Work(ctx, func(_ context.Context, key string) error {
		return errors.New("fail")
	})

	ks.Send(ctx, "a")

	err1 := ks.Close()
	err2 := ks.Close()

	if err1 == nil {
		t.Error("first Close should return error")
	}
	if err2 != nil {
		t.Errorf("second Close should return nil, got %v", err2)
	}
}

func TestKeyShardWorker_ConcurrentSend(t *testing.T) {
	ks, err := NewKeyShardWorker(4, 100)
	if err != nil {
		t.Fatalf("NewKeyShardWorker: %v", err)
	}

	var processed atomic.Int32
	ctx := context.Background()

	ks.Work(ctx, func(_ context.Context, key string) error {
		processed.Add(1)
		return nil
	})

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ks.Send(ctx, fmt.Sprintf("key-%d", i))
		}(i)
	}
	wg.Wait()

	ks.Close()

	if count := processed.Load(); count != 100 {
		t.Errorf("expected 100 processed, got %d", count)
	}
}
