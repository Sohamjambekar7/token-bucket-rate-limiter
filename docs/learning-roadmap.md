# Five-phase learning roadmap

## Phase 1 — In-memory algorithm and concurrency
Start with `internal/store/memory.go`. Read `Allow` from top to bottom. The mutex protects the whole read/refill/decision/write sequence. The lazy refill calculation is performed only when a request arrives:
`newTokens = min(capacity, oldTokens + max(0, now-last) * rate)`.
Try removing the mutex locally and run `go test -race ./...`; the race detector may expose unsafe access, while the concurrent correctness test expresses the business invariant. Do not commit the intentionally broken version.

## Phase 2 — Redis and atomic Lua
Set `STORE=redis` and run Redis. `internal/store/redis.go` uses Lua. A Redis script is atomic with respect to other Redis commands on that server: no other request can interleave between HMGET and HSET. Compare this with separate GET and SET calls, which allow two clients to read the same token count before either writes.
For production, use Redis TIME inside the script instead of trusting each app instance's clock, and consider a Redis connection pool rather than dialing per request (the go-redis client supplies pooling).

## Phase 3 — Sliding log and config
Set a client's strategy to `sliding_log`. The sorted set contains accepted timestamps. Old timestamps are removed before checking the limit. The configuration store validates updates and uses an RWMutex so concurrent reads can proceed together.
The included config store is local to one process; for multiple API replicas, move it to a shared database or Redis and define cache invalidation.

## Phase 4 — HTTP contract
`POST /v1/rate-limit/check` expects `{"client_id":"demo-client"}`. `PUT /v1/admin/clients/{client_id}` accepts `strategy`, `rate`, `burst`, and `window_seconds`. The admin route is intentionally unauthenticated for local learning only; never expose it publicly without authentication and authorization.

## Phase 5 — load and correctness
Run `go test -race ./...` before running k6. The test starts 500 goroutines against one bucket and expects only the initial 25 tokens to pass. k6 generates 500 request arrivals per second. Check the custom ALLOW/DENY counts and the latency percentiles. A load test cannot prove correctness by itself; combine it with deterministic tests and state invariants.
