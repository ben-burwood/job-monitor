package jobs

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTTLCacheCollapsesConcurrentFetches(t *testing.T) {
	c := ttlCache[int]{ttl: time.Minute}
	var calls atomic.Int32
	release := make(chan struct{})
	fetch := func(context.Context) (int, error) {
		calls.Add(1)
		<-release
		return 42, nil
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if v, err := c.get(t.Context(), "id", false, fetch); v != 42 || err != nil {
				t.Errorf("got (%d, %v), want (42, nil)", v, err)
			}
		})
	}
	time.Sleep(50 * time.Millisecond) // let every caller join the in-flight fetch
	close(release)
	wg.Wait()

	if n := calls.Load(); n != 1 {
		t.Errorf("expected 1 fetch, got %d", n)
	}
	if v, _ := c.get(t.Context(), "id", false, fetch); v != 42 || calls.Load() != 1 {
		t.Errorf("expected cached value without refetching")
	}
}

func TestTTLCacheDoesNotCacheErrors(t *testing.T) {
	c := ttlCache[int]{ttl: time.Minute}
	if _, err := c.get(t.Context(), "id", false, func(context.Context) (int, error) { return 0, errors.New("boom") }); err == nil {
		t.Fatal("expected error")
	}
	if v, err := c.get(t.Context(), "id", false, func(context.Context) (int, error) { return 7, nil }); v != 7 || err != nil {
		t.Errorf("got (%d, %v), want (7, nil)", v, err)
	}
}

func TestTTLCacheFetchIgnoresCallerCancellation(t *testing.T) {
	c := ttlCache[error]{ttl: time.Minute}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	got, _ := c.get(ctx, "id", false, func(ctx context.Context) (error, error) { return ctx.Err(), nil })
	if got != nil {
		t.Errorf("fetch saw cancelled context: %v", got)
	}
}
