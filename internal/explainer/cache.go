package explainer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// memoryCache implements an in-memory cache with TTL
type memoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	stats   *CacheStats
	ttl     time.Duration
}

// cacheEntry represents a cached explanation with metadata
type cacheEntry struct {
	explanation *Explanation
	expiresAt   time.Time
	accessCount int64
	lastAccess  time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(ttl time.Duration) Cache {
	cache := &memoryCache{
		entries: make(map[string]*cacheEntry),
		stats: &CacheStats{
			Hits:   0,
			Misses: 0,
		},
		ttl: ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves an explanation from cache
func (c *memoryCache) Get(ctx context.Context, key string) (*Explanation, error) {
	startTime := time.Now()
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.recordMiss()
		return nil, &ExplanationError{
			Code:    ErrCodeCacheError,
			Message: "cache miss",
		}
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		c.recordMiss()
		return nil, &ExplanationError{
			Code:    ErrCodeCacheError,
			Message: "cache entry expired",
		}
	}

	// Update access metadata
	c.mu.Lock()
	entry.accessCount++
	entry.lastAccess = time.Now()
	c.mu.Unlock()

	c.recordHit(time.Since(startTime))
	return entry.explanation, nil
}

// Set stores an explanation in cache
func (c *memoryCache) Set(ctx context.Context, key string, explanation *Explanation, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		explanation: explanation,
		expiresAt:   time.Now().Add(ttl),
		accessCount: 0,
		lastAccess:  time.Now(),
	}

	c.stats.Size = int64(len(c.entries))
	return nil
}

// Delete removes an explanation from cache
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	c.stats.Size = int64(len(c.entries))
	return nil
}

// Clear clears all cached explanations
func (c *memoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
	c.stats.Size = 0
	c.stats.Evictions++
	return nil
}

// Stats returns cache statistics
func (c *memoryCache) Stats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}
	stats.Size = int64(len(c.entries))
	return &stats
}

// Close closes the cache
func (c *memoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = nil
	return nil
}

// recordHit records a cache hit
func (c *memoryCache) recordHit(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.Hits++

	// Update average get time
	totalTime := c.stats.AvgGetTime * (c.stats.Hits - 1)
	c.stats.AvgGetTime = (totalTime + duration.Milliseconds()) / c.stats.Hits

	// Estimate cost savings (assuming $0.002 per explanation)
	c.stats.TotalSaved += 0.002
}

// recordMiss records a cache miss
func (c *memoryCache) recordMiss() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.Misses++
}

// cleanupExpired removes expired entries periodically
func (c *memoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.expiresAt) {
				delete(c.entries, key)
				c.stats.Evictions++
			}
		}
		c.stats.Size = int64(len(c.entries))
		c.mu.Unlock()
	}
}

// GenerateCacheKey generates a cache key for a finding
func GenerateCacheKey(finding interface{}) string {
	// Serialize the finding to JSON
	data, err := json.Marshal(finding)
	if err != nil {
		// Fallback to string representation
		return fmt.Sprintf("%v", finding)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GenerateCacheKeyFromRequest generates a cache key from an explanation request
func GenerateCacheKeyFromRequest(req *ExplanationRequest) string {
	if req == nil || req.Finding == nil {
		return ""
	}

	// Create a deterministic key based on finding attributes
	key := fmt.Sprintf("%s:%s:%s:%s:%d",
		req.Finding.Tool,
		req.Finding.Severity,
		req.Finding.Description,
		req.Finding.File,
		req.Finding.Line,
	)

	// Add CWE if present
	if req.Finding.CWE != "" {
		key += ":" + req.Finding.CWE
	}

	// Hash the key
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CacheMiddleware wraps an explainer with caching
type CacheMiddleware struct {
	explainer Explainer
	cache     Cache
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(explainer Explainer, cache Cache) Explainer {
	return &CacheMiddleware{
		explainer: explainer,
		cache:     cache,
	}
}

// Explain generates an explanation with caching
func (m *CacheMiddleware) Explain(ctx context.Context, req *ExplanationRequest) (*Explanation, error) {
	// Generate cache key
	cacheKey := GenerateCacheKeyFromRequest(req)

	// Try to get from cache
	if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
		cached.Source = "cache"
		cached.CacheKey = cacheKey
		return cached, nil
	}

	// Cache miss - generate new explanation
	explanation, err := m.explainer.Explain(ctx, req)
	if err != nil {
		return nil, err
	}

	// Store in cache
	explanation.CacheKey = cacheKey
	if err := m.cache.Set(ctx, cacheKey, explanation, 0); err != nil {
		// Log error but don't fail the request
		// In production, use proper logging
	}

	return explanation, nil
}

// ExplainBatch generates explanations for multiple vulnerabilities with caching
func (m *CacheMiddleware) ExplainBatch(ctx context.Context, reqs []*ExplanationRequest) ([]*Explanation, error) {
	results := make([]*Explanation, len(reqs))
	uncachedReqs := make([]*ExplanationRequest, 0)
	uncachedIndices := make([]int, 0)

	// Check cache for each request
	for i, req := range reqs {
		cacheKey := GenerateCacheKeyFromRequest(req)
		if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
			cached.Source = "cache"
			cached.CacheKey = cacheKey
			results[i] = cached
		} else {
			uncachedReqs = append(uncachedReqs, req)
			uncachedIndices = append(uncachedIndices, i)
		}
	}

	// Generate explanations for uncached requests
	if len(uncachedReqs) > 0 {
		explanations, err := m.explainer.ExplainBatch(ctx, uncachedReqs)
		if err != nil {
			return nil, err
		}

		// Store in cache and populate results
		for i, explanation := range explanations {
			idx := uncachedIndices[i]
			cacheKey := GenerateCacheKeyFromRequest(uncachedReqs[i])
			explanation.CacheKey = cacheKey

			if err := m.cache.Set(ctx, cacheKey, explanation, 0); err != nil {
				// Log error but continue
			}

			results[idx] = explanation
		}
	}

	return results, nil
}

// GetCacheStats returns cache statistics
func (m *CacheMiddleware) GetCacheStats() *CacheStats {
	return m.cache.Stats()
}

// ClearCache clears the explanation cache
func (m *CacheMiddleware) ClearCache() error {
	return m.cache.Clear(context.Background())
}

// Close closes the explainer and cache
func (m *CacheMiddleware) Close() error {
	if err := m.cache.Close(); err != nil {
		return err
	}
	return m.explainer.Close()
}

// Made with Bob
