package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/token-bucket-rate-limiter/internal/limiter"
)

type handler struct {
	engine  *limiter.Engine
	configs *limiter.ConfigStore
}

func Register(m *http.ServeMux, e *limiter.Engine, c *limiter.ConfigStore) {
	h := handler{engine: e, configs: c}
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	m.HandleFunc("/v1/rate-limit/check", h.check)
	m.HandleFunc("/v1/admin/clients/", h.update)
}
func (h handler) check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		ClientID string `json:"client_id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}
	res, err := h.engine.Check(r.Context(), req.ClientID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Set("X-RateLimit-Limit", itoa(res.Limit))
	w.Header().Set("X-RateLimit-Remaining", itoa(res.Remaining))
	reset := int(res.ResetAfter.Seconds())
	if res.ResetAfter > 0 && reset == 0 {
		reset = 1
	}
	w.Header().Set("X-RateLimit-Reset", itoa(reset))
	status := http.StatusOK
	decision := "ALLOW"
	if !res.Allowed {
		status = http.StatusTooManyRequests
		decision = "DENY"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"decision": decision, "allowed": res.Allowed, "limit": res.Limit, "remaining": res.Remaining})
}
func (h handler) update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", 405)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/admin/clients/")
	var req struct {
		Strategy      string  `json:"strategy"`
		Rate          float64 `json:"rate"`
		Burst         int     `json:"burst"`
		WindowSeconds int     `json:"window_seconds"`
	}
	if id == "" || json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	c := limiter.Config{Strategy: req.Strategy, Rate: req.Rate, Burst: req.Burst, Window: time.Duration(req.WindowSeconds) * time.Second}
	if err := h.configs.Set(id, c); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func itoa(n int) string { return strconv.Itoa(n) }
