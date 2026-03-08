package parallel

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

var ErrClosed = errors.New("keyshard: closed")

type KeyShardWorker struct {
	workerCnt int
	ch        []chan string
	wg        sync.WaitGroup
	closeOnce sync.Once
	started   atomic.Bool
	closed    atomic.Bool
	errs      []error
	errsMu    sync.Mutex
}

func NewKeyShardWorker(workerCnt int, chBuf int) (*KeyShardWorker, error) {
	if workerCnt <= 0 {
		return nil, fmt.Errorf("keyshard: workerCnt must be positive, got %d", workerCnt)
	}
	ch := make([]chan string, workerCnt)
	for i := range workerCnt {
		ch[i] = make(chan string, chBuf)
	}
	return &KeyShardWorker{
		workerCnt: workerCnt,
		ch:        ch,
	}, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func (ksw *KeyShardWorker) Send(ctx context.Context, key string) (err error) {
	if ksw.closed.Load() {
		return ErrClosed
	}

	defer func() {
		if r := recover(); r != nil {
			err = ErrClosed
		}
	}()

	idx := hash(key) % uint32(ksw.workerCnt)
	select {
	case ksw.ch[idx] <- key:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ksw *KeyShardWorker) Work(ctx context.Context, fn func(context.Context, string) error) {
	if !ksw.started.CompareAndSwap(false, true) {
		return
	}
	for i := range ksw.workerCnt {
		ksw.wg.Add(1)
		go func(i int) {
			defer ksw.wg.Done()
			for {
				select {
				case key, ok := <-ksw.ch[i]:
					if !ok {
						return
					}
					if err := fn(ctx, key); err != nil {
						ksw.errsMu.Lock()
						ksw.errs = append(ksw.errs, fmt.Errorf("key %s: %w", key, err))
						ksw.errsMu.Unlock()
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
}

func (ksw *KeyShardWorker) Close() error {
	var err error
	ksw.closeOnce.Do(func() {
		ksw.closed.Store(true)
		for i := range ksw.workerCnt {
			close(ksw.ch[i])
		}
		ksw.wg.Wait()
		ksw.errsMu.Lock()
		err = errors.Join(ksw.errs...)
		ksw.errsMu.Unlock()
	})
	return err
}
