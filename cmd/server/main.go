package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/token-bucket-rate-limiter/internal/limiter"
	"github.com/example/token-bucket-rate-limiter/internal/server"
	"github.com/example/token-bucket-rate-limiter/internal/store"
	"github.com/redis/go-redis/v9"
)

func main() {
	mode := strings.ToLower(env("STORE", "memory"))
	strategy := strings.ToLower(env("STRATEGY", "token_bucket"))
	var st limiter.Store
	if mode == "redis" {
		addr := env("REDIS_ADDR", "localhost:6379")
		client := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("REDIS_PASSWORD")})
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			log.Fatalf("redis ping: %v", err)
		}
		st = store.NewRedis(client)
	} else {
		st = store.NewMemory()
	}
	cfg := limiter.NewConfigStore()
	defaultCfg := limiter.Config{Strategy: strategy, Rate: 5, Burst: 10, Window: time.Second}
	cfg.Set("demo-client", defaultCfg)
	engine := limiter.NewEngine(st, cfg)
	mux := http.NewServeMux()
	server.Register(mux, engine, cfg)
	addr := env("ADDR", ":8080")
	log.Printf("rate limiter listening on %s (store=%s strategy=%s)", addr, mode, strategy)
	log.Fatal(http.ListenAndServe(addr, mux))
}
func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func _unusedAtoi(s string) int { n, _ := strconv.Atoi(s); return n }
