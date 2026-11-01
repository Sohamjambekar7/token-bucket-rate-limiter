package limiter

import (
	"context"
	"time"
)

type Config struct {
	Strategy string        `json:"strategy"`
	Rate     float64       `json:"rate"`  // tokens per second for token bucket; events per window for sliding log
	Burst    int           `json:"burst"` // bucket capacity; maximum events for sliding log
	Window   time.Duration `json:"-"`
}
type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAfter time.Duration
}
type Store interface {
	Allow(context.Context, string, Config, time.Time) (Result, error)
}
