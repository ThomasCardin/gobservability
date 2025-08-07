package pkg

import "github.com/ThomasCardin/peek/shared/types"

// Generate fake data for development with realistic metrics
func generateFakePods(nodeName string) []*types.Pod {
	fakePods := []*types.Pod{
		{
			Name:        "nginx-deployment-abc123",
			ContainerID: "docker://1234567890abcdef",
			PID:         1234,
			PodMetrics: types.PodMetrics{
				CPU: types.PodCPUStats{
					UTime:      12340,
					STime:      5670,
					CPUPercent: 2.5,
				},
				Memory: types.PodMemoryStats{
					VmSize:     512000, // 512MB
					VmRSS:      128000, // 128MB
					MemPercent: 3.2,
				},
				Network: types.PodNetworkStats{
					BytesReceived:    1024000, // 1MB
					BytesTransmitted: 2048000, // 2MB
				},
				Disk: types.PodDiskStats{
					ReadBytes:  5120000, // 5MB
					WriteBytes: 2560000, // 2.5MB
				},
			},
			PidDetails: types.PidDetails{
				Name:                "nginx",
				State:               "S",
				Priority:            20,
				Nice:                0,
				Threads:             4,
				CUTime:              1200,
				CSTime:              800,
				TaskCPU:             0,
				VmPeak:              600000,
				PacketsReceived:     500,
				PacketsTransmitted:  800,
			},
		},
		{
			Name:        "redis-server-xyz789",
			ContainerID: "containerd://fedcba0987654321",
			PID:         5678,
			PodMetrics: types.PodMetrics{
				CPU: types.PodCPUStats{
					UTime:      23450,
					STime:      8900,
					CPUPercent: 5.2,
				},
				Memory: types.PodMemoryStats{
					VmSize:     256000, // 256MB
					VmRSS:      200000, // 200MB
					MemPercent: 5.0,
				},
				Network: types.PodNetworkStats{
					BytesReceived:    512000,  // 512KB
					BytesTransmitted: 1024000, // 1MB
				},
				Disk: types.PodDiskStats{
					ReadBytes:  10240000, // 10MB
					WriteBytes: 7680000,  // 7.5MB
				},
			},
			PidDetails: types.PidDetails{
				Name:                "redis-server",
				State:               "S",
				Priority:            20,
				Nice:                0,
				Threads:             6,
				CUTime:              2300,
				CSTime:              1100,
				TaskCPU:             1,
				VmPeak:              300000,
				PacketsReceived:     200,
				PacketsTransmitted:  300,
			},
		},
		{
			Name:        "api-service-def456",
			ContainerID: "docker://abcdef1234567890",
			PID:         9012,
			PodMetrics: types.PodMetrics{
				CPU: types.PodCPUStats{
					UTime:      45600,
					STime:      12300,
					CPUPercent: 8.7,
				},
				Memory: types.PodMemoryStats{
					VmSize:     1024000, // 1GB
					VmRSS:      768000,  // 768MB
					MemPercent: 19.2,
				},
				Network: types.PodNetworkStats{
					BytesReceived:    5120000,  // 5MB
					BytesTransmitted: 10240000, // 10MB
				},
				Disk: types.PodDiskStats{
					ReadBytes:  20480000, // 20MB
					WriteBytes: 15360000, // 15MB
				},
			},
			PidDetails: types.PidDetails{
				Name:                "node",
				State:               "S",
				Priority:            20,
				Nice:                0,
				Threads:             12,
				CUTime:              4500,
				CSTime:              2800,
				TaskCPU:             2,
				VmPeak:              1200000,
				PacketsReceived:     2500,
				PacketsTransmitted:  4800,
			},
		},
		{
			Name:        "postgres-db-ghi789",
			ContainerID: "containerd://567890abcdef1234",
			PID:         3456,
			PodMetrics: types.PodMetrics{
				CPU: types.PodCPUStats{
					UTime:      67800,
					STime:      23400,
					CPUPercent: 12.3,
				},
				Memory: types.PodMemoryStats{
					VmSize:     2048000, // 2GB
					VmRSS:      1536000, // 1.5GB
					MemPercent: 38.4,
				},
				Network: types.PodNetworkStats{
					BytesReceived:    2048000, // 2MB
					BytesTransmitted: 1024000, // 1MB
				},
				Disk: types.PodDiskStats{
					ReadBytes:  102400000, // 100MB
					WriteBytes: 51200000,  // 50MB
				},
			},
			PidDetails: types.PidDetails{
				Name:                "postgres",
				State:               "S",
				Priority:            20,
				Nice:                0,
				Threads:             8,
				CUTime:              6700,
				CSTime:              4200,
				TaskCPU:             3,
				VmPeak:              2200000,
				PacketsReceived:     1000,
				PacketsTransmitted:  600,
			},
		},
		{
			Name:        "failing-pod-error",
			ContainerID: "Not found",
			PID:         -1,
			PodMetrics:  types.PodMetrics{}, // Empty metrics for failed pod
			PidDetails:  types.PidDetails{}, // Empty details for failed pod
		},
		{
			Name:        "partial-pod-test",
			ContainerID: "docker://errorcontainer123",
			PID:         -1,
			PodMetrics:  types.PodMetrics{}, // Empty metrics for failed pod
			PidDetails:  types.PidDetails{}, // Empty details for failed pod
		},
	}

	return fakePods
}
