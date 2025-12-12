package metrics

import (
	"fmt"
	"time"

	"github.com/ThomasCardin/gobservability/shared/types"
	gocache "github.com/patrickmn/go-cache"
)

type Cache struct {
	nodeCache *gocache.Cache
	podCache  *gocache.Cache
}

type CachedNodeMetrics struct {
	CPU       *types.CPUStats
	Network   *types.NetworkStats
	Disk      *types.DiskStats
	Timestamp time.Time
}

type CachedPodMetrics struct {
	CPU       *types.PodCPUStats
	Network   *types.PodNetworkStats
	Disk      *types.PodDiskStats
	Timestamp time.Time
}

/*
NewCache creates a new cache instance
- Node cache: keep for 1 minute, cleanup every 30 seconds
- Pod cache: keep for 1 minute, cleanup every 30 seconds
*/
func NewCache() *Cache {
	return &Cache{
		nodeCache: gocache.New(1*time.Minute, 30*time.Second),
		podCache:  gocache.New(1*time.Minute, 30*time.Second),
	}
}

// UpdateNodeMetrics stores current node metrics and returns previous values
func (c *Cache) UpdateNodeMetrics(nodeName string, cpu *types.CPUStats, network *types.NetworkStats, disk *types.DiskStats) (*CachedNodeMetrics, bool) {
	key := fmt.Sprintf("node:%s", nodeName)

	// Get previous metrics
	var prev *CachedNodeMetrics
	if cached, found := c.nodeCache.Get(key); found {
		prev = cached.(*CachedNodeMetrics)
	}

	// Store new metrics
	newMetrics := &CachedNodeMetrics{
		CPU:       cpu,
		Network:   network,
		Disk:      disk,
		Timestamp: time.Now(),
	}
	c.nodeCache.Set(key, newMetrics, gocache.DefaultExpiration)

	return prev, prev != nil
}

// UpdatePodMetrics stores current pod metrics and returns previous values
func (c *Cache) UpdatePodMetrics(pid int, cpu *types.PodCPUStats, network *types.PodNetworkStats, disk *types.PodDiskStats) (*CachedPodMetrics, bool) {
	key := fmt.Sprintf("pod:%d", pid)

	// Get previous metrics
	var prev *CachedPodMetrics
	if cached, found := c.podCache.Get(key); found {
		prev = cached.(*CachedPodMetrics)
	}

	// Store new metrics
	newMetrics := &CachedPodMetrics{
		CPU:       cpu,
		Network:   network,
		Disk:      disk,
		Timestamp: time.Now(),
	}
	c.podCache.Set(key, newMetrics, gocache.DefaultExpiration)

	return prev, prev != nil
}

// CleanupStaleEntries removes cache entries for PIDs that no longer exist
func (c *Cache) CleanupStaleEntries(activePIDs []int) {
	// Create a map of active PIDs for quick lookup
	activeMap := make(map[string]bool)
	for _, pid := range activePIDs {
		key := fmt.Sprintf("pod:%d", pid)
		activeMap[key] = true
	}

	// Get all items in pod cache
	items := c.podCache.Items()
	for key := range items {
		if !activeMap[key] {
			c.podCache.Delete(key)
		}
	}
}
