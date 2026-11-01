package server

import (
	"github.com/example/token-bucket-rate-limiter/internal/limiter"
	"github.com/example/token-bucket-rate-limiter/internal/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckEndpointHeadersAndDeny(t *testing.T) {
	cfg := limiter.NewConfigStore()
	_ = cfg.Set("c", limiter.Config{Strategy: "token_bucket", Rate: 0.01, Burst: 1, Window: time.Second})
	e := limiter.NewEngine(store.NewMemory(), cfg)
	mux := http.NewServeMux()
	Register(mux, e, cfg)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/v1/rate-limit/check", strings.NewReader(`{"client_id":"c"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if i == 0 && w.Code != 200 {
			t.Fatal(w.Code)
		}
		if i == 1 && w.Code != 429 {
			t.Fatal(w.Code)
		}
		if w.Header().Get("X-RateLimit-Limit") == "" {
			t.Fatal("missing headers")
		}
	}
}
