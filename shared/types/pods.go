package types

type Pod struct {
	Name            string          `json:"name"`
	ContainerID     string          `json:"container_id"`
	PID             int             `json:"pid"`
	PodMetrics      PodMetrics      `json:"pod_metrics"`
	PidDetails      PidDetails      `json:"pid_details"`
	ResourceLimits  ResourceInfo    `json:"resource_limits"`
	ResourceRequests ResourceInfo   `json:"resource_requests"`
}

// PodMetrics contains only the metrics needed for UI calculations
type PodMetrics struct {
	CPU     PodCPUStats     `json:"cpu"`
	Memory  PodMemoryStats  `json:"memory"`
	Network PodNetworkStats `json:"network"`
	Disk    PodDiskStats    `json:"disk"`
}

// PodCPUStats contains only CPU metrics used by CalculateUIPod
type PodCPUStats struct {
	UTime      uint64  `json:"utime"`       // User mode jiffies
	STime      uint64  `json:"stime"`       // Kernel mode jiffies
	CPUPercent float64 `json:"cpu_percent"` // Calculated CPU %
}

// PodMemoryStats contains only memory metrics used by CalculateUIPod
type PodMemoryStats struct {
	VmSize     uint64  `json:"vm_size"`     // Virtual memory size (KB)
	VmRSS      uint64  `json:"vm_rss"`      // Resident memory size (KB)
	MemPercent float64 `json:"mem_percent"` // % of total node memory
}

// PodNetworkStats contains only network metrics used by CalculateUIPod
type PodNetworkStats struct {
	BytesReceived    uint64 `json:"bytes_received"`
	BytesTransmitted uint64 `json:"bytes_transmitted"`
}

// PodDiskStats contains only disk metrics used by CalculateUIPod
type PodDiskStats struct {
	ReadBytes  uint64 `json:"read_bytes"`  // Bytes read from disk
	WriteBytes uint64 `json:"write_bytes"` // Bytes written to disk
}

// ResourceInfo contains resource limits and requests for a pod
type ResourceInfo struct {
	CPU    string `json:"cpu"`    // CPU in millicores (e.g., "100m", "2")
	Memory string `json:"memory"` // Memory in bytes (e.g., "128Mi", "1Gi")
}
