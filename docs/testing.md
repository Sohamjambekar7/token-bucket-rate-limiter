# Test and load-test guide

## Unit and race tests
```bash
go test ./...
go test -race ./...
```
The race detector checks unsafe concurrent memory access. `TestConcurrentNoDoubleSpend` starts 500 goroutines against a bucket with 25 initial tokens and asserts exactly 25 successes.

## k6
Start the API, then:
```bash
k6 run scripts/load.js
```
The script targets 500 arrivals/second for 20 seconds and records ALLOW/DENY counters. The rate is a request arrival rate, not necessarily 500 simultaneously active VUs. k6 reports latency percentiles and dropped iterations.

## What correctness means
For a bucket initialized with B tokens, no refill during the test, and N concurrent checks, the number of ALLOW decisions must be `min(B, N)`. With refill, expected allows also include only tokens mathematically accrued up to capacity. Validate decisions and state invariants—not just HTTP status or throughput.

## Benchmark expectations
Do not assume sub-millisecond service latency. Run on the target machine, record p50/p95/p99, CPU, memory, Redis RTT, and error rates. The pure in-memory algorithm can be microsecond-scale in favorable conditions; HTTP and distributed Redis paths have additional overhead.
