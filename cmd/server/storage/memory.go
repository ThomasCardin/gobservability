package storage

import (
	"time"

	"github.com/ThomasCardin/peek/shared/types"
	"github.com/patrickmn/go-cache"
)

var GlobalStore = NewCacheStore(10*time.Second, 5*time.Second)

type CacheStore struct {
	cache *cache.Cache
}

func NewCacheStore(defaultExpiration, cleanupInterval time.Duration) *CacheStore {
	return &CacheStore{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// StoreNodeStats stores incoming node statistics from agents
func (s *CacheStore) StoreNodeStats(stats types.NodeStatsPayload) {
	s.cache.Set(stats.NodeName, stats, cache.DefaultExpiration)
}

// GetAllNodes returns all stored node statistics
func (s *CacheStore) GetAllNodes() map[string]*types.NodeStatsPayload {
	result := make(map[string]*types.NodeStatsPayload)

	for key, item := range s.cache.Items() {
		if stats, ok := item.Object.(types.NodeStatsPayload); ok {
			result[key] = &stats
		}
	}

	return result
}

// GetNodeStats returns statistics for a specific node
func (s *CacheStore) GetNodeStats(nodeName string) (*types.NodeStatsPayload, bool) {
	if item, found := s.cache.Get(nodeName); found {
		if stats, ok := item.(types.NodeStatsPayload); ok {
			return &stats, true
		}
	}
	return nil, false
}

// GetCacheStats returns cache statistics for monitoring
func (s *CacheStore) GetCacheStats() (int, int) {
	return s.cache.ItemCount(), len(s.cache.Items())
}