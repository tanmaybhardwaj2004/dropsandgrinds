package config

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback to treating it as just an address if parsing fails
		opt = &redis.Options{
			Addr:     redisURL,
			Password: "", // no password set
			DB:       0,  // use default DB
		}
	}
	client := redis.NewClient(opt)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s", redisURL)
	return client, nil
}
