package fbhttp

import (
	"context"
	"fmt"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

const uploadCacheTTL = 3 * time.Minute

// UploadCache is an interface for tracking active uploads.
// Allows for different backends (e.g. in-memory or redis)
// to support both single instance and multi replica deployments.
type UploadCache interface {
	// Register stores an upload with its expected file size. remove is called if
	// the upload expires before completion, to delete the partial file; it must
	// route through the uploading user's scoped filesystem so that eviction
	// cannot follow a symlink out of the user's scope.
	Register(filePath string, fileSize int64, remove func() error)

	// Complete removes an upload from the cache
	Complete(filePath string)

	// GetLength returns the expected file size for an active upload
	GetLength(filePath string) (int64, error)

	// Touch refreshes the TTL for an active upload
	Touch(filePath string)

	// Close cleans up any resources
	Close()
}

// memoryUploadEntry is the value stored for each active upload.
type memoryUploadEntry struct {
	size   int64
	remove func() error
}

// memoryUploadCache is an upload cache for single replica deployments
type memoryUploadCache struct {
	cache *ttlcache.Cache[string, memoryUploadEntry]
}

func newMemoryUploadCache() *memoryUploadCache {
	cache := ttlcache.New[string, memoryUploadEntry]()
	cache.OnEviction(func(_ context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, memoryUploadEntry]) {
		if reason == ttlcache.EvictionReasonExpired {
			fmt.Printf("deleting incomplete upload file: \"%s\"\n", item.Key())
			// Delete through the scoped removal callback rather than a raw
			// os.Remove on the cached path, so an ancestor directory swapped for
			// a symlink during the TTL window cannot redirect the delete outside
			// the user's scope.
			if remove := item.Value().remove; remove != nil {
				if err := remove(); err != nil {
					fmt.Printf("failed to delete incomplete upload file %q: %v\n", item.Key(), err)
				}
			}
		}
	})
	go cache.Start()

	return &memoryUploadCache{cache: cache}
}

func (c *memoryUploadCache) Register(filePath string, fileSize int64, remove func() error) {
	c.cache.Set(filePath, memoryUploadEntry{size: fileSize, remove: remove}, uploadCacheTTL)
}

func (c *memoryUploadCache) Complete(filePath string) {
	c.cache.Delete(filePath)
}

func (c *memoryUploadCache) GetLength(filePath string) (int64, error) {
	item := c.cache.Get(filePath)
	if item == nil {
		return 0, fmt.Errorf("no active upload found for the given path")
	}
	return item.Value().size, nil
}

func (c *memoryUploadCache) Touch(filePath string) {
	c.cache.Touch(filePath)
}

func (c *memoryUploadCache) Close() {
	c.cache.Stop()
}

// NewUploadCache creates a new upload cache.
// If redisURL is empty, an in-memory cache will be used (suitable for single instance deployments).
// Otherwise, Redis will be used for the cache (suitable for multi-instance deployments).
// The redisURL can include credentials, e.g. redis://user:pass@host:port
func NewUploadCache(redisURL string) (UploadCache, error) {
	if redisURL != "" {
		return newRedisUploadCache(redisURL)
	}
	return newMemoryUploadCache(), nil
}
