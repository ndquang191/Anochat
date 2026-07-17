package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

local values = redis.call("HMGET", key, "tokens", "updated_at")
local tokens = tonumber(values[1])
local updatedAt = tonumber(values[2])

if tokens == nil or updatedAt == nil then
    tokens = burst
    updatedAt = now
end

local elapsed = math.max(0, now - updatedAt)
tokens = math.min(burst, tokens + (elapsed * rate / 1000))

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "updated_at", now)
redis.call("PEXPIRE", key, ttl)
return allowed
`)

func RateLimitMiddleware(rdb *redis.Client, requestsPerSecond, burst int) gin.HandlerFunc {
	// Keep idle buckets only long enough to refill completely twice.
	bucketTTL := time.Duration(2*burst) * time.Second / time.Duration(requestsPerSecond)
	if bucketTTL < time.Second {
		bucketTTL = time.Second
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:%s", ip)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()

		allowed, err := tokenBucketScript.Run(
			ctx,
			rdb,
			[]string{key},
			requestsPerSecond,
			burst,
			time.Now().UnixMilli(),
			bucketTTL.Milliseconds(),
		).Int()
		if err != nil {
			slog.Warn("Rate limiter Redis error, allowing request", "error", err, "ip", ip)
			c.Next()
			return
		}

		if allowed == 0 {
			slog.Warn("Rate limit exceeded", "ip", ip, "rate", requestsPerSecond, "burst", burst)
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
