# Edge cases and operational checklist

1. **Clock moves backwards:** clamp negative elapsed time to zero. Prefer a monotonic clock in one process; distributed services need a consistent time source. Redis TIME inside Lua is a strong production option.
2. **Long idle period:** cap at capacity; never accumulate unlimited tokens.
3. **Same timestamp:** Sliding Log members must be unique even when multiple requests share one millisecond. The implementation adds a nanosecond suffix.
4. **Window boundary:** remove scores `<= now-window`; requests exactly at the boundary are outside the active window.
5. **Client key missing/unknown:** reject as a client error; production policy may choose a default config instead.
6. **Config changed while requests run:** each request reads one config snapshot. Existing token balance may exceed a newly lowered capacity until the next evaluation clamps it.
7. **Redis unavailable:** decide fail-open or fail-closed explicitly. This sample returns an API error; it does not silently allow traffic.
8. **Redis key expiry:** TTL should outlive the refill horizon; otherwise idle state can expire and recreate a full bucket. The sample uses a conservative TTL.
9. **Hot client key:** a single Redis key can become a serialization hotspot. Consider local pre-limits, key partitioning only if semantics allow, or dedicated capacity.
10. **Admin security:** authenticate, authorize, validate, audit, and rate-limit admin updates.
11. **Spoofed identity:** derive client ID from a verified API key/JWT or trusted gateway header.
12. **Retries:** a retried check consumes another token. If clients need idempotency, add a request ID and deduplication policy.
13. **Sliding log memory:** cap active clients, set TTL, and monitor ZSET cardinality.
14. **Multi-region:** a single Redis primary is not a globally linearizable multi-region limiter. Define whether approximate regional quotas are acceptable.
15. **Response headers:** Reset is expressed as seconds until the next token/request can be accepted, rounded up for a useful retry hint.
16. **Latency:** benchmark local algorithm, HTTP service, and Redis path separately. A networked Redis decision cannot generally guarantee sub-millisecond end-to-end latency.
