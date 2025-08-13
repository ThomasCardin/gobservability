package grpc

import (
	pb "github.com/ThomasCardin/gobservability/proto"
	"github.com/ThomasCardin/gobservability/shared/types"
)

// Conversions from Go types to gRPC protobuf (for agent -> server)

func ConvertToGRPCMetrics(metrics types.NodeMetrics) *pb.NodeMetrics {
	return &pb.NodeMetrics{
		Cpu:     ConvertToGRPCCPUStats(metrics.CPU),
		Memory:  ConvertToGRPCMemoryStats(metrics.Memory),
		Network: ConvertToGRPCNetworkStats(metrics.Network),
		Disk:    ConvertToGRPCDiskStats(metrics.Disk),
		Pods:    ConvertToGRPCPods(metrics.Pods),
	}
}

func ConvertToGRPCCPUStats(cpu *types.CPUStats) *pb.CPUStats {
	if cpu == nil {
		return nil
	}
	return &pb.CPUStats{
		User:       int64(cpu.User),
		Nice:       int64(cpu.Nice),
		System:     int64(cpu.System),
		Idle:       int64(cpu.Idle),
		Iowait:     int64(cpu.IOWait),
		Irq:        int64(cpu.IRQ),
		Softirq:    int64(cpu.SoftIRQ),
		Steal:      int64(cpu.Steal),
		Total:      int64(cpu.Total),
		CpuPercent: cpu.CPUPercent,
	}
}

func ConvertToGRPCMemoryStats(mem *types.MemoryStats) *pb.MemoryStats {
	if mem == nil {
		return nil
	}
	return &pb.MemoryStats{
		MemTotal:      int64(mem.MemTotal),
		MemFree:       int64(mem.MemFree),
		MemAvailable:  int64(mem.MemAvailable),
		Buffers:       int64(mem.Buffers),
		Cached:        int64(mem.Cached),
		SwapCached:    int64(mem.SwapCached),
		SwapTotal:     int64(mem.SwapTotal),
		SwapFree:      int64(mem.SwapFree),
		MemoryPercent: mem.MemoryPercent,
	}
}

func ConvertToGRPCNetworkStats(net *types.NetworkStats) *pb.NetworkStats {
	if net == nil {
		return nil
	}
	return &pb.NetworkStats{
		BytesReceived:      net.BytesReceived,
		BytesTransmitted:   net.BytesTransmitted,
		PacketsReceived:    net.PacketsReceived,
		PacketsTransmitted: net.PacketsTransmitted,
		ErrorsReceived:     net.ErrorsReceived,
		ErrorsTransmitted:  net.ErrorsTransmitted,
		RxRate:             net.RxRate,
		TxRate:             net.TxRate,
		TotalRate:          net.TotalRate,
	}
}

func ConvertToGRPCDiskStats(disk *types.DiskStats) *pb.DiskStats {
	if disk == nil {
		return nil
	}
	return &pb.DiskStats{
		ReadsCompleted:  disk.ReadsCompleted,
		ReadsMerged:     disk.ReadsMerged,
		SectorsRead:     disk.SectorsRead,
		TimeReading:     disk.TimeReading,
		WritesCompleted: disk.WritesCompleted,
		WritesMerged:    disk.WritesMerged,
		SectorsWritten:  disk.SectorsWritten,
		TimeWriting:     disk.TimeWriting,
		ReadRate:        disk.ReadRate,
		WriteRate:       disk.WriteRate,
		TotalRate:       disk.TotalRate,
	}
}

func ConvertToGRPCPods(pods []*types.Pod) []*pb.Pod {
	grpcPods := make([]*pb.Pod, len(pods))
	for i, pod := range pods {
		grpcPods[i] = &pb.Pod{
			Name:        pod.Name,
			ContainerId: pod.ContainerID,
			Pid:         int64(pod.PID),
			PodMetrics:  ConvertToGRPCPodMetrics(pod.PodMetrics),
			PidDetails:  ConvertToGRPCPidDetails(pod.PidDetails),
		}
	}
	return grpcPods
}

func ConvertToGRPCPodMetrics(metrics types.PodMetrics) *pb.PodMetrics {
	return &pb.PodMetrics{
		Cpu:     ConvertToGRPCPodCPUStats(metrics.CPU),
		Memory:  ConvertToGRPCPodMemoryStats(metrics.Memory),
		Network: ConvertToGRPCPodNetworkStats(metrics.Network),
		Disk:    ConvertToGRPCPodDiskStats(metrics.Disk),
	}
}

func ConvertToGRPCPodCPUStats(cpu types.PodCPUStats) *pb.PodCPUStats {
	return &pb.PodCPUStats{
		Utime:      cpu.UTime,
		Stime:      cpu.STime,
		CpuPercent: cpu.CPUPercent,
	}
}

func ConvertToGRPCPodMemoryStats(mem types.PodMemoryStats) *pb.PodMemoryStats {
	return &pb.PodMemoryStats{
		VmSize:     mem.VmSize,
		VmRss:      mem.VmRSS,
		MemPercent: mem.MemPercent,
	}
}

func ConvertToGRPCPodNetworkStats(net types.PodNetworkStats) *pb.PodNetworkStats {
	return &pb.PodNetworkStats{
		BytesReceived:    net.BytesReceived,
		BytesTransmitted: net.BytesTransmitted,
	}
}

func ConvertToGRPCPodDiskStats(disk types.PodDiskStats) *pb.PodDiskStats {
	return &pb.PodDiskStats{
		ReadBytes:  disk.ReadBytes,
		WriteBytes: disk.WriteBytes,
	}
}

func ConvertToGRPCPidDetails(pid types.PidDetails) *pb.PidDetails {
	return &pb.PidDetails{
		// From /proc/{PID}/stat - process basics
		Name:             pid.Name,
		State:            pid.State,
		Priority:         int32(pid.Priority),
		Nice:             int32(pid.Nice),
		Threads:          int32(pid.Threads),
		StartTime:        pid.StartTime,
		RealtimePriority: int32(pid.RealtimePriority),
		Cutime:           pid.CUTime,
		Cstime:           pid.CSTime,
		TaskCpu:          int32(pid.TaskCPU),

		// From /proc/{PID}/status - security and scheduling
		Kthread:                   int32(pid.KThread),
		Seccomp:                   int32(pid.Seccomp),
		SpeculationIndirectBranch: pid.SpeculationIndirectBranch,
		CpusAllowedList:           pid.CpusAllowedList,
		MemsAllowedList:           pid.MemsAllowedList,
		VoluntaryCtxtSwitches:     pid.VoluntaryCtxtSwitches,
		NonvoluntaryCtxtSwitches:  pid.NonvoluntaryCtxtSwitches,
		VmPeak:                    pid.VmPeak,
		VmLck:                     pid.VmLck,
		VmPin:                     pid.VmPin,

		// From /proc/{PID}/net - network
		PacketsReceived:    pid.PacketsReceived,
		PacketsTransmitted: pid.PacketsTransmitted,
		ErrorsReceived:     pid.ErrorsReceived,
		ErrorsTransmitted:  pid.ErrorsTransmitted,

		// From /proc/{PID}/io
		CancelledWrites: pid.CancelledWrites,

		// Additional process information
		Cmdline: pid.Cmdline,
		Stack:   pid.Stack,
		OpenFds: int32(pid.OpenFDs),
		MaxFds:  pid.MaxFDs,
		Cgroup:  pid.Cgroup,
		VmData:  pid.VmData,
		VmStk:   pid.VmStk,
		VmExe:   pid.VmExe,
		VmLib:   pid.VmLib,
		VmSwap:  pid.VmSwap,
	}
}

// Conversions from gRPC protobuf to Go types (for server <- agent)

func ConvertNodeMetrics(grpcMetrics *pb.NodeMetrics) types.NodeMetrics {
	return types.NodeMetrics{
		CPU:     ConvertCPUStats(grpcMetrics.Cpu),
		Memory:  ConvertMemoryStats(grpcMetrics.Memory),
		Network: ConvertNetworkStats(grpcMetrics.Network),
		Disk:    ConvertDiskStats(grpcMetrics.Disk),
		Pods:    ConvertPods(grpcMetrics.Pods),
	}
}

func ConvertCPUStats(grpc *pb.CPUStats) *types.CPUStats {
	if grpc == nil {
		return nil
	}
	return &types.CPUStats{
		User:       int(grpc.User),
		Nice:       int(grpc.Nice),
		System:     int(grpc.System),
		Idle:       int(grpc.Idle),
		IOWait:     int(grpc.Iowait),
		IRQ:        int(grpc.Irq),
		SoftIRQ:    int(grpc.Softirq),
		Steal:      int(grpc.Steal),
		Total:      int(grpc.Total),
		CPUPercent: grpc.CpuPercent,
	}
}

func ConvertMemoryStats(grpc *pb.MemoryStats) *types.MemoryStats {
	if grpc == nil {
		return nil
	}
	return &types.MemoryStats{
		MemTotal:      int(grpc.MemTotal),
		MemFree:       int(grpc.MemFree),
		MemAvailable:  int(grpc.MemAvailable),
		Buffers:       int(grpc.Buffers),
		Cached:        int(grpc.Cached),
		SwapCached:    int(grpc.SwapCached),
		SwapTotal:     int(grpc.SwapTotal),
		SwapFree:      int(grpc.SwapFree),
		MemoryPercent: grpc.MemoryPercent,
	}
}

func ConvertNetworkStats(grpc *pb.NetworkStats) *types.NetworkStats {
	if grpc == nil {
		return nil
	}
	return &types.NetworkStats{
		BytesReceived:      grpc.BytesReceived,
		BytesTransmitted:   grpc.BytesTransmitted,
		PacketsReceived:    grpc.PacketsReceived,
		PacketsTransmitted: grpc.PacketsTransmitted,
		ErrorsReceived:     grpc.ErrorsReceived,
		ErrorsTransmitted:  grpc.ErrorsTransmitted,
		RxRate:             grpc.RxRate,
		TxRate:             grpc.TxRate,
		TotalRate:          grpc.TotalRate,
	}
}

func ConvertDiskStats(grpc *pb.DiskStats) *types.DiskStats {
	if grpc == nil {
		return nil
	}
	return &types.DiskStats{
		ReadsCompleted:  grpc.ReadsCompleted,
		ReadsMerged:     grpc.ReadsMerged,
		SectorsRead:     grpc.SectorsRead,
		TimeReading:     grpc.TimeReading,
		WritesCompleted: grpc.WritesCompleted,
		WritesMerged:    grpc.WritesMerged,
		SectorsWritten:  grpc.SectorsWritten,
		TimeWriting:     grpc.TimeWriting,
		ReadRate:        grpc.ReadRate,
		WriteRate:       grpc.WriteRate,
		TotalRate:       grpc.TotalRate,
	}
}

func ConvertPods(grpcPods []*pb.Pod) []*types.Pod {
	pods := make([]*types.Pod, len(grpcPods))
	for i, grpcPod := range grpcPods {
		pods[i] = &types.Pod{
			Name:        grpcPod.Name,
			ContainerID: grpcPod.ContainerId,
			PID:         int(grpcPod.Pid),
			PodMetrics:  ConvertPodMetrics(grpcPod.PodMetrics),
			PidDetails:  ConvertPidDetails(grpcPod.PidDetails),
		}
	}
	return pods
}

func ConvertPodMetrics(grpc *pb.PodMetrics) types.PodMetrics {
	if grpc == nil {
		return types.PodMetrics{}
	}
	return types.PodMetrics{
		CPU:     ConvertPodCPUStats(grpc.Cpu),
		Memory:  ConvertPodMemoryStats(grpc.Memory),
		Network: ConvertPodNetworkStats(grpc.Network),
		Disk:    ConvertPodDiskStats(grpc.Disk),
	}
}

func ConvertPodCPUStats(grpc *pb.PodCPUStats) types.PodCPUStats {
	if grpc == nil {
		return types.PodCPUStats{}
	}
	return types.PodCPUStats{
		UTime:      grpc.Utime,
		STime:      grpc.Stime,
		CPUPercent: grpc.CpuPercent,
	}
}

func ConvertPodMemoryStats(grpc *pb.PodMemoryStats) types.PodMemoryStats {
	if grpc == nil {
		return types.PodMemoryStats{}
	}
	return types.PodMemoryStats{
		VmSize:     grpc.VmSize,
		VmRSS:      grpc.VmRss,
		MemPercent: grpc.MemPercent,
	}
}

func ConvertPodNetworkStats(grpc *pb.PodNetworkStats) types.PodNetworkStats {
	if grpc == nil {
		return types.PodNetworkStats{}
	}
	return types.PodNetworkStats{
		BytesReceived:    grpc.BytesReceived,
		BytesTransmitted: grpc.BytesTransmitted,
	}
}

func ConvertPodDiskStats(grpc *pb.PodDiskStats) types.PodDiskStats {
	if grpc == nil {
		return types.PodDiskStats{}
	}
	return types.PodDiskStats{
		ReadBytes:  grpc.ReadBytes,
		WriteBytes: grpc.WriteBytes,
	}
}

func ConvertPidDetails(grpc *pb.PidDetails) types.PidDetails {
	if grpc == nil {
		return types.PidDetails{}
	}
	return types.PidDetails{
		// From /proc/{PID}/stat - process basics
		Name:             grpc.Name,
		State:            grpc.State,
		Priority:         int(grpc.Priority),
		Nice:             int(grpc.Nice),
		Threads:          int(grpc.Threads),
		StartTime:        grpc.StartTime,
		RealtimePriority: int(grpc.RealtimePriority),
		CUTime:           grpc.Cutime,
		CSTime:           grpc.Cstime,
		TaskCPU:          int(grpc.TaskCpu),

		// From /proc/{PID}/status - security and scheduling
		KThread:                   int(grpc.Kthread),
		Seccomp:                   int(grpc.Seccomp),
		SpeculationIndirectBranch: grpc.SpeculationIndirectBranch,
		CpusAllowedList:           grpc.CpusAllowedList,
		MemsAllowedList:           grpc.MemsAllowedList,
		VoluntaryCtxtSwitches:     grpc.VoluntaryCtxtSwitches,
		NonvoluntaryCtxtSwitches:  grpc.NonvoluntaryCtxtSwitches,
		VmPeak:                    grpc.VmPeak,
		VmLck:                     grpc.VmLck,
		VmPin:                     grpc.VmPin,

		// From /proc/{PID}/net - network
		PacketsReceived:    grpc.PacketsReceived,
		PacketsTransmitted: grpc.PacketsTransmitted,
		ErrorsReceived:     grpc.ErrorsReceived,
		ErrorsTransmitted:  grpc.ErrorsTransmitted,

		// From /proc/{PID}/io
		CancelledWrites: grpc.CancelledWrites,

		// Additional process information
		Cmdline: grpc.Cmdline,
		Stack:   grpc.Stack,
		OpenFDs: int(grpc.OpenFds),
		MaxFDs:  grpc.MaxFds,
		Cgroup:  grpc.Cgroup,
		VmData:  grpc.VmData,
		VmStk:   grpc.VmStk,
		VmExe:   grpc.VmExe,
		VmLib:   grpc.VmLib,
		VmSwap:  grpc.VmSwap,
	}
}
