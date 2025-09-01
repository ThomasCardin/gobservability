package storage

import (
	"time"

	"github.com/ThomasCardin/gobservability/cmd/server/alerts"
	pb "github.com/ThomasCardin/gobservability/proto"
	"github.com/ThomasCardin/gobservability/shared/types"
	"github.com/patrickmn/go-cache"
)

var GlobalStore = NewCacheStore(10*time.Second, 5*time.Second)

type CacheStore struct {
	cache           *cache.Cache
	flamegraphTasks *cache.Cache
	alertsManager   *alerts.AlertsManager
}

type FlamegraphTask struct {
	TaskID    string
	NodeName  string
	PodName   string
	Format    string
	Completed bool
	Error     string
	Data      []byte
	CreatedAt time.Time
}

func NewCacheStore(defaultExpiration, cleanupInterval time.Duration) *CacheStore {
	return &CacheStore{
		cache:           cache.New(defaultExpiration, cleanupInterval),
		flamegraphTasks: cache.New(30*time.Minute, 5*time.Minute), // Tasks expire after 30 minutes
		alertsManager:   nil, // Set later via SetAlertsManager
	}
}

// SetAlertsManager sets the alerts manager for metric evaluation
func (s *CacheStore) SetAlertsManager(manager *alerts.AlertsManager) {
	s.alertsManager = manager
}

// StoreNodeStats stores incoming node statistics from agents
func (s *CacheStore) StoreNodeStats(stats types.NodeStatsPayload) {
	s.cache.Set(stats.NodeName, stats, cache.DefaultExpiration)
	
	// Evaluate alerts if alerts manager is available
	if s.alertsManager != nil {
		s.alertsManager.EvaluateMetrics(stats)
	}
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

// StoreFlamegraphResult stores the result of a flamegraph generation task
func (s *CacheStore) StoreFlamegraphResult(taskID string, resp *pb.FlamegraphResponse, err error, format, nodeName, podName string) {
	task := &FlamegraphTask{
		TaskID:    taskID,
		NodeName:  nodeName,
		PodName:   podName,
		Format:    format,
		Completed: true,
		CreatedAt: time.Now(),
	}

	if err != nil {
		task.Error = err.Error()
	} else if resp != nil {
		if resp.Error != "" {
			task.Error = resp.Error
		} else {
			task.Data = resp.FlamegraphData
		}
	}

	s.flamegraphTasks.Set(taskID, task, cache.DefaultExpiration)
}

// GetFlamegraphResult retrieves the result of a flamegraph generation task
func (s *CacheStore) GetFlamegraphResult(taskID string) *FlamegraphTask {
	if item, found := s.flamegraphTasks.Get(taskID); found {
		if task, ok := item.(*FlamegraphTask); ok {
			return task
		}
	}
	return nil
}

// CreateFlamegraphTask creates a new flamegraph task in pending state
func (s *CacheStore) CreateFlamegraphTask(taskID, nodeName, podName, format string) {
	task := &FlamegraphTask{
		TaskID:    taskID,
		NodeName:  nodeName,
		PodName:   podName,
		Format:    format,
		Completed: false,
		CreatedAt: time.Now(),
	}

	s.flamegraphTasks.Set(taskID, task, cache.DefaultExpiration)
}
