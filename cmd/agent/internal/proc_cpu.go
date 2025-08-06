package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ThomasCardin/peek/shared/types"
)

const (
	PROC_STAT = "/proc/stat"
)

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#17-miscellaneous-kernel-statistics-in-procstat
func ProcStat() (*types.CPUStats, error) {
	file, err := os.Open(PROC_STAT)
	if err != nil {
		return nil, fmt.Errorf("error: opening %s %v", PROC_STAT, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return nil, fmt.Errorf("error: invalid %s format", PROC_STAT)
			}

			user, _ := strconv.ParseUint(fields[1], 10, 64)
			nice, _ := strconv.ParseUint(fields[2], 10, 64)
			system, _ := strconv.ParseUint(fields[3], 10, 64)
			idle, _ := strconv.ParseUint(fields[4], 10, 64)
			iowait, _ := strconv.ParseUint(fields[5], 10, 64)
			irq, _ := strconv.ParseUint(fields[6], 10, 64)
			softirq, _ := strconv.ParseUint(fields[7], 10, 64)

			return &types.CPUStats{
				User:    int(user),
				Nice:    int(nice),
				System:  int(system),
				Idle:    int(idle),
				IOWait:  int(iowait),
				IRQ:     int(irq),
				SoftIRQ: int(softirq),
				Total:   int(user + nice + system + idle + iowait + irq + softirq),
			}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: reading %s: %v", PROC_STAT, err)
	}

	return nil, fmt.Errorf("error: finding cpu ")
}
