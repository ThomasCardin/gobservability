package metrics

import (
	"time"

	"github.com/ThomasCardin/gobservability/shared/types"
)

// NodeCache stores previous metrics for delta calculations
type NodeCache struct {
	LastNodeCPU       *types.CPUStats
	LastNodeNetwork   *types.NetworkStats
	LastNodeDisk      *types.DiskStats
	LastNodeTimestamp time.Time

	PodCache map[int]*PodCacheEntry
}

type PodCacheEntry struct {
	LastCPU       *types.PodCPUStats
	LastNetwork   *types.PodNetworkStats
	LastDisk      *types.PodDiskStats
	LastTimestamp time.Time
}

// NewMetricsCache creates a new metrics cache
func NewMetricsCache() *NodeCache {
	return &NodeCache{
		PodCache: make(map[int]*PodCacheEntry),
	}
}
