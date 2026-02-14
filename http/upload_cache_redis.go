package fbhttp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// redisUploadCache is an upload cache for multi replica deployments
type redisUploadCache struct {
	client *redis.Client
}

func newRedisUploadCache(redisURL string) (*redisUploadCache, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &redisUploadCache{client: client}, nil
}

func (c *redisUploadCache) filePathKey(filePath string) string {
	return "filebrowser:upload:" + filePath
}

func (c *redisUploadCache) Register(filePath string, fileSize int64) {
	err := c.client.Set(context.Background(), c.filePathKey(filePath), fileSize, uploadCacheTTL).Err()
	if err != nil {
		log.Printf("failed to register upload in redis cache: %v", err)
	}
}

func (c *redisUploadCache) Complete(filePath string) {
	err := c.client.Del(context.Background(), c.filePathKey(filePath)).Err()
	if err != nil {
		log.Printf("failed to complete upload in redis cache: %v", err)
	}
}

func (c *redisUploadCache) GetLength(filePath string) (int64, error) {
	result, err := c.client.Get(context.Background(), c.filePathKey(filePath)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, fmt.Errorf("no active upload found for the given path")
		}
		return 0, fmt.Errorf("redis error: %w", err)
	}

	size, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload length in cache: %w", err)
	}

	return size, nil
}

func (c *redisUploadCache) Touch(filePath string) {
	err := c.client.Expire(context.Background(), c.filePathKey(filePath), uploadCacheTTL).Err()
	if err != nil {
		log.Printf("failed to touch upload in redis cache: %v", err)
	}
}

func (c *redisUploadCache) Close() {
	c.client.Close()
}
