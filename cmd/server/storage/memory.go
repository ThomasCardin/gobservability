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

func (s *CacheStore) StoreNodeStats(stats types.NodeStatsPayload) {
	s.cache.Set(stats.NodeName, stats, cache.DefaultExpiration)
}

func (s *CacheStore) GetAllNodes() map[string]*types.NodeStatsPayload {
	result := make(map[string]*types.NodeStatsPayload)

	for key, item := range s.cache.Items() {
		if stats, ok := item.Object.(types.NodeStatsPayload); ok {
			result[key] = &stats
		}
	}

	return result
}

func (s *CacheStore) GetNodeStats(nodeName string) (*types.NodeStatsPayload, bool) {
	if item, found := s.cache.Get(nodeName); found {
		if stats, ok := item.(types.NodeStatsPayload); ok {
			return &stats, true
		}
	}
	return nil, false
}

func (s *CacheStore) GetCacheStats() (int, int) {
	return s.cache.ItemCount(), len(s.cache.Items())
}

// UINode cache methods
func (s *CacheStore) StoreUINode(nodeName string, uiNode interface{}) {
	s.cache.Set("ui_"+nodeName, uiNode, cache.DefaultExpiration)
}

func (s *CacheStore) GetUINode(nodeName string) (interface{}, bool) {
	return s.cache.Get("ui_" + nodeName)
}

func (s *CacheStore) GetAllUINodes() map[string]interface{} {
	result := make(map[string]interface{})

	for key, item := range s.cache.Items() {
		if len(key) > 3 && key[:3] == "ui_" {
			nodeName := key[3:]
			result[nodeName] = item.Object
		}
	}

	return result
}
