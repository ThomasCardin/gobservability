package types

import (
	"time"
)

type NodeMetrics struct {
	CPU     *CPUStats     `json:"cpu"`
	Memory  *MemoryStats  `json:"memory"`
	Network *NetworkStats `json:"network"`
	Disk    *DiskStats    `json:"disk"`
	Pods    []*Pod        `json:"pods"`
}

type NodeStatsPayload struct {
	NodeName  string      `json:"node_name"`
	Timestamp time.Time   `json:"timestamp"`
	Metrics   NodeMetrics `json:"metrics"`
}

