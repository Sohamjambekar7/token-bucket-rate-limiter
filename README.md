# Standalone Rate Limiter Service

A learning-oriented Go implementation of a standalone HTTP rate-limiter with an in-memory store and an optional Redis store. It includes Token Bucket and Sliding Window Log strategies, dynamic per-client configuration, tests, Docker Compose, and a k6 load script.

## Requirements
- Go 1.22+
- Redis 7+ only for distributed mode
- Docker Compose (optional)
- k6 (optional load test)

## Run in memory
```bash
go test -race ./...
go run ./cmd/server
```
Default client: `demo-client` (5 tokens/sec, burst 10).

## Run with Redis
```bash
docker compose up --build
```
This starts Redis and the API with `STORE=redis`.

## API
Check a request:
```bash
curl -i -X POST http://localhost:8080/v1/rate-limit/check   -H 'Content-Type: application/json'   -d '{"client_id":"demo-client"}'
```
- `200` means ALLOW.
- `429` means DENY.
- Every decision includes `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` (seconds until retry; zero when no wait is currently required).

Update a client:
```bash
curl -i -X PUT http://localhost:8080/v1/admin/clients/mobile-app   -H 'Content-Type: application/json'   -d '{"strategy":"token_bucket","rate":3,"burst":8,"window_seconds":60}'
```
Use `strategy: "sliding_log"` to select Sliding Window Log. For that strategy, `burst` is the maximum events allowed in `window_seconds`; `rate` must be positive for validation but is not used by the log algorithm.

## Important production notes
- The sample admin endpoint has no authentication. Put it behind authenticated admin access before exposing it.
- The sample config store is process-local. For multi-replica deployments, distribute config through a database/config service or Redis and define cache invalidation.
- Redis Lua makes each decision atomic on one Redis primary. It does not make Redis highly available by itself; persistence, failover, replication, and clock behavior need operational design.
- Use a trusted client identity. Do not blindly trust a client-supplied ID in public deployments.
- The Redis scripts use server-supplied application time in this teaching version. Production deployments should consider Redis TIME inside Lua to avoid clock skew between API replicas.
- The in-memory implementation serializes all keys with one mutex for clarity. A sharded-lock or per-key lock design can improve throughput.
- Sub-millisecond end-to-end latency is not a safe promise. Network, scheduler, TLS, JSON, and Redis round-trip costs dominate. Measure p50/p95/p99 on your hardware; benchmark the algorithm separately.
