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

func getProcStat(devMode string) string {
	return shared.GetProcBasePath(devMode) + "/stat"
}

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#17-miscellaneous-kernel-statistics-in-procstat
func ProcStat(devMode string) (*types.CPUStats, error) {
	procStatPath := getProcStat(devMode)
	file, err := os.Open(procStatPath)
	if err != nil {
		return nil, fmt.Errorf("error: opening %s %v", procStatPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return nil, fmt.Errorf("error: invalid %s format", procStatPath)
			}

			user, _ := strconv.ParseUint(fields[1], 10, 64)
			nice, _ := strconv.ParseUint(fields[2], 10, 64)
			system, _ := strconv.ParseUint(fields[3], 10, 64)
			idle, _ := strconv.ParseUint(fields[4], 10, 64)
			iowait, _ := strconv.ParseUint(fields[5], 10, 64)
			irq, _ := strconv.ParseUint(fields[6], 10, 64)
			softirq, _ := strconv.ParseUint(fields[7], 10, 64)
			steal := uint64(0)
			if len(fields) >= 9 {
				steal, _ = strconv.ParseUint(fields[8], 10, 64)
			}

			return &types.CPUStats{
				User:    int(user),
				Nice:    int(nice),
				System:  int(system),
				Idle:    int(idle),
				IOWait:  int(iowait),
				IRQ:     int(irq),
				SoftIRQ: int(softirq),
				Steal:   int(steal),
				Total:   int(user + nice + system + idle + iowait + irq + softirq + steal),
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: reading %s: %v", procStatPath, err)
	}

	return nil, fmt.Errorf("error: finding cpu ")
}
