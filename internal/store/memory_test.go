package store

import (
	"context"
	"github.com/example/token-bucket-rate-limiter/internal/limiter"
	"sync"
	"testing"
	"time"
)

func TestTokenBucketLazyRefillAndCapacity(t *testing.T) {
	m := NewMemory()
	c := limiter.Config{Strategy: "token_bucket", Rate: 2, Burst: 2}
	now := time.Unix(100, 0)
	for i := 0; i < 2; i++ {
		r, _ := m.Allow(context.Background(), "a", c, now)
		if !r.Allowed {
			t.Fatal("initial burst should pass")
		}
	}
	r, _ := m.Allow(context.Background(), "a", c, now)
	if r.Allowed {
		t.Fatal("bucket should be empty")
	}
	r, _ = m.Allow(context.Background(), "a", c, now.Add(500*time.Millisecond))
	if !r.Allowed {
		t.Fatal("one token should refill after 500ms")
	}
	r, _ = m.Allow(context.Background(), "a", c, now.Add(10*time.Second))
	if r.Remaining > 2 {
		t.Fatal("capacity must cap tokens")
	}
}
func TestConcurrentNoDoubleSpend(t *testing.T) {
	m := NewMemory()
	c := limiter.Config{Strategy: "token_bucket", Rate: 0.000001, Burst: 25}
	now := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, _ := m.Allow(context.Background(), "same", c, now)
			if r.Allowed {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 25 {
		t.Fatalf("got %d allows, want exactly 25", allowed)
	}
}
