package kubernetes

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
				Name:               "nginx",
				State:              "S",
				Priority:           20,
				Nice:               0,
				Threads:            4,
				CUTime:             1200,
				CSTime:             800,
				TaskCPU:            0,
				VmPeak:             600000,
				PacketsReceived:    500,
				PacketsTransmitted: 800,
				Cmdline:            "/usr/sbin/nginx -g daemon off;",
				OpenFDs:            15,
				MaxFDs:             1024,
				Cgroup: []string{
					"12:perf_event:/kubepods/besteffort/pod123abc",
					"11:hugetlb:/kubepods/besteffort/pod123abc",
					"10:pids:/kubepods/besteffort/pod123abc",
					"9:freezer:/kubepods/besteffort/pod123abc",
					"8:memory:/kubepods/besteffort/pod123abc",
					"7:cpu,cpuacct:/kubepods/besteffort/pod123abc",
					"6:devices:/kubepods/besteffort/pod123abc",
					"5:net_cls,net_prio:/kubepods/besteffort/pod123abc",
				},
				Stack: []string{
					"[<0>] __schedule+0x2e0/0x870",
					"[<0>] schedule+0x36/0x80",
					"[<0>] schedule_hrtimeout_range_clock+0x104/0x110",
					"[<0>] poll_schedule_timeout+0x43/0x70",
					"[<0>] do_select+0x789/0x830",
					"[<0>] core_sys_select+0x17c/0x2f0",
					"[<0>] kern_select+0xd1/0x110",
					"[<0>] __x64_sys_select+0x3d/0x50",
					"[<0>] do_syscall_64+0x5c/0x90",
					"[<0>] entry_SYSCALL_64_after_hwframe+0x44/0xae",
				},
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
				Name:               "redis-server",
				State:              "S",
				Priority:           20,
				Nice:               0,
				Threads:            6,
				CUTime:             2300,
				CSTime:             1100,
				TaskCPU:            1,
				VmPeak:             300000,
				PacketsReceived:    200,
				PacketsTransmitted: 300,
				Cmdline:            "/usr/local/bin/redis-server *:6379",
				OpenFDs:            8,
				MaxFDs:             10240,
				Cgroup: []string{
					"12:perf_event:/kubepods/burstable/podxyz789",
					"11:hugetlb:/kubepods/burstable/podxyz789",
					"10:pids:/kubepods/burstable/podxyz789",
					"9:freezer:/kubepods/burstable/podxyz789",
					"8:memory:/kubepods/burstable/podxyz789",
					"7:cpu,cpuacct:/kubepods/burstable/podxyz789",
					"6:devices:/kubepods/burstable/podxyz789",
				},
				Stack: []string{
					"[<0>] ep_poll+0x4c7/0x4e0",
					"[<0>] do_epoll_wait+0xb0/0xd0",
					"[<0>] __x64_sys_epoll_wait+0x1a/0x20",
					"[<0>] do_syscall_64+0x5c/0x90",
					"[<0>] entry_SYSCALL_64_after_hwframe+0x44/0xae",
				},
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
				Name:               "node",
				State:              "S",
				Priority:           20,
				Nice:               0,
				Threads:            12,
				CUTime:             4500,
				CSTime:             2800,
				TaskCPU:            2,
				VmPeak:             1200000,
				PacketsReceived:    2500,
				PacketsTransmitted: 4800,
				Cmdline:            "node /app/server.js --port=3000 --env=production",
				OpenFDs:            45,
				MaxFDs:             65536,
				Cgroup: []string{
					"12:perf_event:/kubepods/guaranteed/pod456def",
					"11:hugetlb:/kubepods/guaranteed/pod456def",
					"10:pids:/kubepods/guaranteed/pod456def/container123",
					"9:freezer:/kubepods/guaranteed/pod456def",
					"8:memory:/kubepods/guaranteed/pod456def/container123",
					"7:cpu,cpuacct:/kubepods/guaranteed/pod456def/container123",
					"6:devices:/kubepods/guaranteed/pod456def",
					"5:net_cls,net_prio:/kubepods/guaranteed/pod456def",
					"4:cpuset:/kubepods/guaranteed/pod456def",
					"3:blkio:/kubepods/guaranteed/pod456def",
				},
				Stack: []string{
					"[<0>] futex_wait_queue_me+0xc5/0x120",
					"[<0>] futex_wait+0x10c/0x250",
					"[<0>] do_futex+0x106/0x5a0",
					"[<0>] __x64_sys_futex+0x13c/0x180",
					"[<0>] do_syscall_64+0x5c/0x90",
					"[<0>] entry_SYSCALL_64_after_hwframe+0x44/0xae",
				},
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
				Name:               "postgres",
				State:              "S",
				Priority:           20,
				Nice:               0,
				Threads:            8,
				CUTime:             6700,
				CSTime:             4200,
				TaskCPU:            3,
				VmPeak:             2200000,
				PacketsReceived:    1000,
				PacketsTransmitted: 600,
				Cmdline:            "postgres: main process",
				OpenFDs:            25,
				MaxFDs:             1024,
				Cgroup: []string{
					"12:perf_event:/kubepods/burstable/pod789ghi/postgres-container",
					"11:hugetlb:/kubepods/burstable/pod789ghi",
					"10:pids:/kubepods/burstable/pod789ghi/postgres-container",
					"9:freezer:/kubepods/burstable/pod789ghi",
					"8:memory:/kubepods/burstable/pod789ghi/postgres-container",
					"7:cpu,cpuacct:/kubepods/burstable/pod789ghi/postgres-container",
					"6:devices:/kubepods/burstable/pod789ghi",
					"5:net_cls,net_prio:/kubepods/burstable/pod789ghi",
					"2:name=systemd:/kubepods/burstable/pod789ghi",
				},
				Stack: []string{
					"[<0>] do_wait+0x1fb/0x2b0",
					"[<0>] kernel_wait4+0x8c/0x140",
					"[<0>] __do_sys_wait4+0x85/0x90",
					"[<0>] __x64_sys_wait4+0x1f/0x30",
					"[<0>] do_syscall_64+0x5c/0x90",
					"[<0>] entry_SYSCALL_64_after_hwframe+0x44/0xae",
				},
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
