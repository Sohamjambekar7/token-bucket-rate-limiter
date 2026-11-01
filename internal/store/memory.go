package store

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/example/token-bucket-rate-limiter/internal/limiter"
)

type bucket struct {
	tokens float64
	last   time.Time
}
type Memory struct {
	mu      sync.Mutex
	buckets map[string]bucket
	events  map[string][]time.Time
}

func NewMemory() *Memory {
	return &Memory{buckets: map[string]bucket{}, events: map[string][]time.Time{}}
}
func (m *Memory) Allow(_ context.Context, key string, c limiter.Config, now time.Time) (limiter.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.Strategy == "sliding_log" {
		return m.sliding(key, c, now), nil
	}
	b, ok := m.buckets[key]
	if !ok {
		b = bucket{tokens: float64(c.Burst), last: now}
	}
	elapsed := now.Sub(b.last).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	b.tokens = math.Min(float64(c.Burst), b.tokens+elapsed*c.Rate)
	b.last = now
	allowed := b.tokens >= 1
	if allowed {
		b.tokens--
	}
	m.buckets[key] = b
	reset := time.Duration(0)
	if !allowed {
		reset = time.Duration((1 - b.tokens) / c.Rate * float64(time.Second))
	}
	return limiter.Result{Allowed: allowed, Limit: c.Burst, Remaining: int(math.Floor(b.tokens)), ResetAfter: reset}, nil
}
func (m *Memory) sliding(key string, c limiter.Config, now time.Time) limiter.Result {
	cutoff := now.Add(-c.Window)
	old := m.events[key]
	kept := old[:0]
	for _, t := range old {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	allowed := len(kept) < c.Burst
	if allowed {
		kept = append(kept, now)
	}
	m.events[key] = kept
	reset := time.Duration(0)
	if len(kept) > 0 {
		reset = kept[0].Add(c.Window).Sub(now)
		if reset < 0 {
			reset = 0
		}
	}
	return limiter.Result{Allowed: allowed, Limit: c.Burst, Remaining: max(0, c.Burst-len(kept)), ResetAfter: reset}
}
