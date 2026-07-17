package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestRateLimitMiddlewareAllowsBurstThenRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	router := gin.New()
	router.Use(RateLimitMiddleware(rdb, 2, 3))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for i := 0; i < 3; i++ {
		response := performRateLimitedRequest(router)
		if response.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i+1, response.Code, http.StatusNoContent)
		}
	}

	response := performRateLimitedRequest(router)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("request after burst status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
	if response.Header().Get("Retry-After") != "1" {
		t.Errorf("Retry-After = %q, want 1", response.Header().Get("Retry-After"))
	}
}

func TestTokenBucketRefillsAtConfiguredRate(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	ctx := context.Background()
	key := "rl:test"
	for i := 0; i < 3; i++ {
		allowed, err := tokenBucketScript.Run(ctx, rdb, []string{key}, 2, 3, 1000, 3000).Int()
		if err != nil {
			t.Fatalf("consume token %d: %v", i+1, err)
		}
		if allowed != 1 {
			t.Fatalf("consume token %d allowed = %d, want 1", i+1, allowed)
		}
	}

	allowed, err := tokenBucketScript.Run(ctx, rdb, []string{key}, 2, 3, 1000, 3000).Int()
	if err != nil {
		t.Fatalf("empty bucket: %v", err)
	}
	if allowed != 0 {
		t.Fatalf("empty bucket allowed = %d, want 0", allowed)
	}

	// At 2 requests/second, 500 ms refills exactly one token.
	allowed, err = tokenBucketScript.Run(ctx, rdb, []string{key}, 2, 3, 1500, 3000).Int()
	if err != nil {
		t.Fatalf("refilled bucket: %v", err)
	}
	if allowed != 1 {
		t.Fatalf("refilled bucket allowed = %d, want 1", allowed)
	}
}

func performRateLimitedRequest(router http.Handler) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "192.0.2.1:1234"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
