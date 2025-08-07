package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ThomasCardin/peek/cmd/agent/shared"
	"github.com/ThomasCardin/peek/shared/types"
)

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#11-process-specific-subdirectories
func ProcPIDStat(devMode string, pid int) (*types.PodCPUStats, *types.PidDetails, error) {
	procPath := fmt.Sprintf("%s/%d/stat", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, nil, fmt.Errorf("error reading %s", procPath)
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 44 { // We need at least 44 fields for all our data
		return nil, nil, fmt.Errorf("insufficient fields in %s", procPath)
	}

	// Parse CPU stats (fields 13, 14, 15, 16)
	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	cutime, _ := strconv.ParseUint(fields[15], 10, 64)
	cstime, _ := strconv.ParseUint(fields[16], 10, 64)
	taskCPU, _ := strconv.Atoi(fields[38])

	cpuStats := &types.PodCPUStats{
		UTime: utime,
		STime: stime,
		// CPUPercent will be calculated later with time delta
	}

	// Parse process details
	name := fields[1]
	// Remove parentheses around name
	if len(name) > 2 && name[0] == '(' && name[len(name)-1] == ')' {
		name = name[1 : len(name)-1]
	}
	state := fields[2]
	priority, _ := strconv.Atoi(fields[17])
	nice, _ := strconv.Atoi(fields[18])
	threads, _ := strconv.Atoi(fields[19])
	startTime, _ := strconv.ParseUint(fields[21], 10, 64)
	realtimePriority, _ := strconv.Atoi(fields[39])

	pidDetails := &types.PidDetails{
		Name:             name,
		State:            state,
		Priority:         priority,
		Nice:             nice,
		Threads:          threads,
		StartTime:        startTime,
		RealtimePriority: realtimePriority,
		CUTime:           cutime,
		CSTime:           cstime,
		TaskCPU:          taskCPU,
	}

	return cpuStats, pidDetails, nil
}

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#11-process-specific-subdirectories
func ProcPIDStatus(devMode string, pid int) (*types.PodMemoryStats, *types.PidDetails, error) {
	procPath := fmt.Sprintf("%s/%d/status", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	memStats := &types.PodMemoryStats{}
	pidDetails := &types.PidDetails{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")

		switch key {
		case "VmPeak":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmPeak = val
			}
		case "VmSize":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memStats.VmSize = val
			}
		case "VmRSS":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memStats.VmRSS = val
			}
		case "VmLck":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmLck = val
			}
		case "VmPin":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmPin = val
			}
		case "Seccomp":
			if val, err := strconv.Atoi(fields[1]); err == nil {
				pidDetails.Seccomp = val
			}
		case "Speculation_Store_Bypass":
			if len(fields) > 1 {
				pidDetails.SpeculationIndirectBranch = strings.Join(fields[1:], " ")
			}
		case "Cpus_allowed_list":
			if len(fields) > 1 {
				pidDetails.CpusAllowedList = fields[1]
			}
		case "Mems_allowed_list":
			if len(fields) > 1 {
				pidDetails.MemsAllowedList = fields[1]
			}
		case "voluntary_ctxt_switches":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VoluntaryCtxtSwitches = val
			}
		case "nonvoluntary_ctxt_switches":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.NonvoluntaryCtxtSwitches = val
			}
		case "VmData":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmData = val
			}
		case "VmStk":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmStk = val
			}
		case "VmExe":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmExe = val
			}
		case "VmLib":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmLib = val
			}
		case "VmSwap":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.VmSwap = val
			}
		}
	}

	return memStats, pidDetails, scanner.Err()
}

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#33--procpidio---display-the-io-accounting-fields
func ProcPIDIO(devMode string, pid int) (*types.PodDiskStats, *types.PidDetails, error) {
	procPath := fmt.Sprintf("%s/%d/io", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	diskStats := &types.PodDiskStats{}
	pidDetails := &types.PidDetails{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")

		switch key {
		case "read_bytes":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				diskStats.ReadBytes = val
			}
		case "write_bytes":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				diskStats.WriteBytes = val
			}
		case "cancelled_write_bytes":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				pidDetails.CancelledWrites = val
			}
		}
	}

	return diskStats, pidDetails, scanner.Err()
}

// ProcPIDNetDev reads /proc/{PID}/net/dev for network stats
func ProcPIDNetDev(devMode string, pid int) (*types.PodNetworkStats, *types.PidDetails, error) {
	procPath := fmt.Sprintf("%s/%d/net/dev", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	netStats := &types.PodNetworkStats{}
	pidDetails := &types.PidDetails{}

	scanner := bufio.NewScanner(file)
	// Skip header lines
	scanner.Scan() // Inter-|   Receive
	scanner.Scan() // face |bytes    packets

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 17 {
			continue
		}

		// Skip loopback interface
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}

		// Parse received stats (fields 1-8)
		if bytesRx, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
			netStats.BytesReceived += bytesRx
		}
		if packetsRx, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
			pidDetails.PacketsReceived += packetsRx
		}
		if errorsRx, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
			pidDetails.ErrorsReceived += errorsRx
		}

		// Parse transmitted stats (fields 9-16)
		if bytesTx, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
			netStats.BytesTransmitted += bytesTx
		}
		if packetsTx, err := strconv.ParseUint(fields[10], 10, 64); err == nil {
			pidDetails.PacketsTransmitted += packetsTx
		}
		if errorsTx, err := strconv.ParseUint(fields[11], 10, 64); err == nil {
			pidDetails.ErrorsTransmitted += errorsTx
		}
	}

	return netStats, pidDetails, scanner.Err()
}

// ProcPIDCmdline reads /proc/{PID}/cmdline
func ProcPIDCmdline(devMode string, pid int) (string, error) {
	procPath := fmt.Sprintf("%s/%d/cmdline", shared.GetProcBasePath(devMode), pid)
	data, err := os.ReadFile(procPath)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %v", procPath, err)
	}

	// Replace null bytes with spaces for display
	cmdline := strings.ReplaceAll(string(data), "\000", " ")
	return strings.TrimSpace(cmdline), nil
}

// ProcPIDStack reads /proc/{PID}/stack
func ProcPIDStack(devMode string, pid int) ([]string, error) {
	procPath := fmt.Sprintf("%s/%d/stack", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	var stack []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			stack = append(stack, line)
		}
	}
	return stack, scanner.Err()
}

// ProcPIDFDCount counts open file descriptors in /proc/{PID}/fd/
func ProcPIDFDCount(devMode string, pid int) (int, error) {
	procPath := fmt.Sprintf("%s/%d/fd", shared.GetProcBasePath(devMode), pid)
	files, err := os.ReadDir(procPath)
	if err != nil {
		return 0, fmt.Errorf("error reading %s: %v", procPath, err)
	}
	return len(files), nil
}

// ProcPIDCgroup reads /proc/{PID}/cgroup
func ProcPIDCgroup(devMode string, pid int) ([]string, error) {
	procPath := fmt.Sprintf("%s/%d/cgroup", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	var cgroups []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			cgroups = append(cgroups, line)
		}
	}
	return cgroups, scanner.Err()
}

// ProcPIDLimits reads /proc/{PID}/limits to get max file descriptors
func ProcPIDLimits(devMode string, pid int) (uint64, error) {
	procPath := fmt.Sprintf("%s/%d/limits", shared.GetProcBasePath(devMode), pid)
	file, err := os.Open(procPath)
	if err != nil {
		return 0, fmt.Errorf("error opening %s: %v", procPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Max open files") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				if maxFiles, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
					return maxFiles, nil
				}
			}
		}
	}
	return 0, scanner.Err()
}

// ProcPIDStatm reads /proc/{PID}/statm for additional memory info
func ProcPIDStatm(devMode string, pid int) (uint64, uint64, uint64, uint64, error) {
	procPath := fmt.Sprintf("%s/%d/statm", shared.GetProcBasePath(devMode), pid)
	data, err := os.ReadFile(procPath)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error reading %s: %v", procPath, err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 7 {
		return 0, 0, 0, 0, fmt.Errorf("insufficient fields in %s", procPath)
	}

	// Convert from pages to KB (assuming 4KB pages)
	pageSize := uint64(4)

	size, _ := strconv.ParseUint(fields[0], 10, 64)     // Total program size
	data_val, _ := strconv.ParseUint(fields[5], 10, 64) // Data segment size
	stk, _ := strconv.ParseUint(fields[6], 10, 64)      // Stack size
	lib := uint64(0)                                    // Library size not directly available in statm

	return size * pageSize, data_val * pageSize, stk * pageSize, lib, nil
}

// CollectPodMetrics combines all PID-based metrics collection
func CollectPodMetrics(devMode string, pid int) (*types.PodMetrics, *types.PidDetails, error) {
	if pid <= 0 {
		return nil, nil, fmt.Errorf("invalid PID: %d", pid)
	}

	// Collect CPU stats and basic process details
	cpuStats, pidDetails1, err := ProcPIDStat(devMode, pid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read stat: %v", err)
	}

	// Collect memory stats and additional process details
	memStats, pidDetails2, err := ProcPIDStatus(devMode, pid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read status: %v", err)
	}

	// Collect disk I/O stats
	diskStats, pidDetails3, err := ProcPIDIO(devMode, pid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read io: %v", err)
	}

	// Collect network stats
	netStats, pidDetails4, err := ProcPIDNetDev(devMode, pid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read net/dev: %v", err)
	}

	// Collect additional process information
	cmdline, _ := ProcPIDCmdline(devMode, pid)
	stack, _ := ProcPIDStack(devMode, pid)
	openFDs, _ := ProcPIDFDCount(devMode, pid)
	maxFDs, _ := ProcPIDLimits(devMode, pid)
	cgroups, _ := ProcPIDCgroup(devMode, pid)

	// Merge pidDetails from all sources
	mergedPidDetails := *pidDetails1
	mergedPidDetails.KThread = pidDetails2.KThread
	mergedPidDetails.Seccomp = pidDetails2.Seccomp
	mergedPidDetails.SpeculationIndirectBranch = pidDetails2.SpeculationIndirectBranch
	mergedPidDetails.CpusAllowedList = pidDetails2.CpusAllowedList
	mergedPidDetails.MemsAllowedList = pidDetails2.MemsAllowedList
	mergedPidDetails.VoluntaryCtxtSwitches = pidDetails2.VoluntaryCtxtSwitches
	mergedPidDetails.NonvoluntaryCtxtSwitches = pidDetails2.NonvoluntaryCtxtSwitches
	// Add detailed metrics from pidDetails2 (memory), pidDetails3 (disk) and pidDetails4 (network)
	mergedPidDetails.VmPeak = pidDetails2.VmPeak
	mergedPidDetails.VmLck = pidDetails2.VmLck
	mergedPidDetails.VmPin = pidDetails2.VmPin
	mergedPidDetails.VmData = pidDetails2.VmData
	mergedPidDetails.VmStk = pidDetails2.VmStk
	mergedPidDetails.VmExe = pidDetails2.VmExe
	mergedPidDetails.VmLib = pidDetails2.VmLib
	mergedPidDetails.VmSwap = pidDetails2.VmSwap
	mergedPidDetails.CancelledWrites = pidDetails3.CancelledWrites
	mergedPidDetails.PacketsReceived = pidDetails4.PacketsReceived
	mergedPidDetails.PacketsTransmitted = pidDetails4.PacketsTransmitted
	mergedPidDetails.ErrorsReceived = pidDetails4.ErrorsReceived
	mergedPidDetails.ErrorsTransmitted = pidDetails4.ErrorsTransmitted

	// Add new process information
	mergedPidDetails.Cmdline = cmdline
	mergedPidDetails.Stack = stack
	mergedPidDetails.OpenFDs = openFDs
	mergedPidDetails.MaxFDs = maxFDs
	mergedPidDetails.Cgroup = cgroups

	podMetrics := &types.PodMetrics{
		CPU:     *cpuStats,
		Memory:  *memStats,
		Network: *netStats,
		Disk:    *diskStats,
	}

	return podMetrics, &mergedPidDetails, nil
}
