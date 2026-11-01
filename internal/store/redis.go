package store

import (
	"context"
	"fmt"
	"time"

	"github.com/example/token-bucket-rate-limiter/internal/limiter"
	"github.com/redis/go-redis/v9"
)

const tokenScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])
local state = redis.call('HMGET', key, 'tokens', 'ts')
local tokens = tonumber(state[1]) or capacity
local last = tonumber(state[2]) or now
local elapsed = math.max(0, now - last)
tokens = math.min(capacity, tokens + elapsed * rate)
local allowed = 0
if tokens >= 1 then tokens = tokens - 1; allowed = 1 end
redis.call('HSET', key, 'tokens', tokens, 'ts', now)
redis.call('PEXPIRE', key, ttl)
local reset = 0
if allowed == 0 then reset = math.ceil((1 - tokens) / rate * 1000) end
return {allowed, math.floor(tokens), reset}
`
const slidingScript = `
local key=KEYS[1]
local now=tonumber(ARGV[1])
local window=tonumber(ARGV[2])
local maxn=tonumber(ARGV[3])
local cutoff=now-window
redis.call('ZREMRANGEBYSCORE',key,'-inf',cutoff)
local count=redis.call('ZCARD',key)
local allowed=0
if count < maxn then
  redis.call('ZADD',key,now,ARGV[4])
  allowed=1
  count=count+1
end
redis.call('PEXPIRE',key,window)
local first=redis.call('ZRANGE',key,0,0,'WITHSCORES')
local reset=0
if #first > 0 then reset=math.max(0,tonumber(first[2])+window-now) end
return {allowed, math.max(0,maxn-count), reset}
`

type Redis struct {
	client  *redis.Client
	token   *redis.Script
	sliding *redis.Script
}

func NewRedis(c *redis.Client) *Redis {
	return &Redis{client: c, token: redis.NewScript(tokenScript), sliding: redis.NewScript(slidingScript)}
}
func (r *Redis) Allow(ctx context.Context, key string, c limiter.Config, now time.Time) (limiter.Result, error) {
	ms := now.UnixMilli()
	var out []interface{}
	var err error
	if c.Strategy == "sliding_log" {
		out, err = r.sliding.Run(ctx, r.client, []string{"rl:sliding:" + key}, ms, c.Window.Milliseconds(), c.Burst, fmt.Sprintf("%d-%d", ms, time.Now().UnixNano())).Slice()
	} else {
		ttl := int64((float64(c.Burst)/c.Rate*3 + 60) * 1000)
		out, err = r.token.Run(ctx, r.client, []string{"rl:bucket:" + key}, c.Burst, c.Rate, float64(ms)/1000, ttl).Slice()
	}
	if err != nil {
		return limiter.Result{}, err
	}
	n := func(i int) int64 { v, _ := redisInt(out[i]); return v }
	return limiter.Result{Allowed: n(0) == 1, Remaining: int(n(1)), Limit: c.Burst, ResetAfter: time.Duration(n(2)) * time.Millisecond}, nil
}
func redisInt(v interface{}) (int64, error) {
	switch x := v.(type) {
	case int64:
		return x, nil
	case string:
		var n int64
		_, e := fmt.Sscan(x, &n)
		return n, e
	default:
		return 0, fmt.Errorf("unexpected redis value %T", v)
	}
}
