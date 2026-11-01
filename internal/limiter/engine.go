package limiter

import (
	"context"
	"errors"
	"time"
)

type Engine struct {
	store   Store
	configs *ConfigStore
}

func NewEngine(s Store, c *ConfigStore) *Engine { return &Engine{store: s, configs: c} }
func (e *Engine) Check(ctx context.Context, id string, now time.Time) (Result, error) {
	if id == "" {
		return Result{}, errors.New("client_id is required")
	}
	c, ok := e.configs.Get(id)
	if !ok {
		return Result{}, errors.New("unknown client_id")
	}
	return e.store.Allow(ctx, id, c, now)
}
