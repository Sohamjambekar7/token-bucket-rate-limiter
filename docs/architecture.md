# Architecture and correctness notes

```text
Client
  |
  | POST /v1/rate-limit/check {client_id}
  v
HTTP API ---- ConfigStore (client strategy/rate/burst/window)
  |
  v
Engine
  |
  +---- Memory Store: mutex + per-client bucket/event log
  |
  +---- Redis Store: one Lua script per decision
             |
             +-- HASH for Token Bucket
             +-- ZSET for Sliding Window Log
```

## Token Bucket state transition
For capacity C, refill rate R tokens/second, prior balance T and elapsed seconds Δt:

`T' = min(C, T + max(0, Δt) × R)`

If `T' >= 1`, allow and set `Tnext = T' - 1`; otherwise deny and keep `Tnext = T'`. The next decision performs refill. No timer thread is needed.

## Why a mutex / Lua script?
A naive `read tokens -> calculate -> write tokens` sequence can interleave:
two goroutines both read 1 token, both decide allow, and both write 0. One token was spent twice. The in-memory mutex makes the entire transition one critical section. Redis Lua runs the read/compute/write script atomically relative to other Redis commands.

## Sliding Window Log
Keep accepted request timestamps in a sorted set. Before deciding, remove timestamps at or before `now - window`. Allow if the remaining count is below the configured maximum; add a unique member timestamp on allow. This is accurate but uses O(requests-in-window) memory per client. A Sliding Window Counter is a cheaper approximation if memory is a concern.

## Complexity
- Token Bucket: O(1) state and O(1) work per decision.
- Sliding Window Log: O(log n) sorted-set insert/removal plus O(log n) range operations; memory O(n) events per active client/window.
