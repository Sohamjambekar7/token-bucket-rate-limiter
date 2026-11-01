package limiter

import (
	"errors"
	"sync"
	"time"
)

type ConfigStore struct {
	mu     sync.RWMutex
	values map[string]Config
}

func NewConfigStore() *ConfigStore { return &ConfigStore{values: make(map[string]Config)} }
func (s *ConfigStore) Set(id string, c Config) error {
	if id == "" || c.Burst < 1 || c.Rate <= 0 {
		return errors.New("client_id, rate > 0 and burst >= 1 are required")
	}
	if c.Strategy != "token_bucket" && c.Strategy != "sliding_log" {
		return errors.New("strategy must be token_bucket or sliding_log")
	}
	if c.Window <= 0 {
		c.Window = time.Second
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[id] = c
	return nil
}
func (s *ConfigStore) Get(id string) (Config, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.values[id]
	return c, ok
}
