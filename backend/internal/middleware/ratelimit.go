package middleware

import (
    "context"
    "fmt"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    redis_rate "github.com/go-redis/redis_rate/v10"
)

var (
	redisClient *redis.Client
	limiter     *redis_rate.Limiter
)

func init() {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", 
		DB:       0,
	})
	limiter = redis_rate.NewLimiter(redisClient)
}

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.Next()
			return
		}

		key := fmt.Sprintf("rl:user:%d", userID)
		res, err := limiter.Allow(context.Background(), key, redis_rate.PerSecond(2))
		if err != nil {
			c.JSON(500, gin.H{"error": "rate limiter error"})
			c.Abort()
			return
		}

		if res.Allowed == 0 {
			c.Header("Retry-After", fmt.Sprintf("%d", int(res.RetryAfter.Seconds())))
			c.JSON(429, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}