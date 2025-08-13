package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ThomasCardin/gobservability/cmd/agent/shared"
	"github.com/ThomasCardin/gobservability/shared/types"
)

func getProcMeminfo(devMode string) string {
	return shared.GetProcBasePath(devMode) + "/meminfo"
}

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#meminfo
func ProcMeminfo(devMode string) (*types.MemoryStats, error) {
	procMeminfoPath := getProcMeminfo(devMode)
	file, err := os.Open(procMeminfoPath)
	if err != nil {
		return nil, fmt.Errorf("error: opening %s %v", procMeminfoPath, err)
	}
	defer file.Close()

	memStats := &types.MemoryStats{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal":
			memStats.MemTotal = int(value)
		case "MemFree":
			memStats.MemFree = int(value)
		case "MemAvailable":
			memStats.MemAvailable = int(value)
		case "Buffers":
			memStats.Buffers = int(value)
		case "Cached":
			memStats.Cached = int(value)
		case "SwapCached":
			memStats.SwapCached = int(value)
		case "SwapTotal":
			memStats.SwapTotal = int(value)
		case "SwapFree":
			memStats.SwapFree = int(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: reading %s: %v", procMeminfoPath, err)
	}

	return memStats, nil
}
