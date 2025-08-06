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
	PROC_DISKSTATS = "/proc/diskstats"
)

// https://github.com/torvalds/linux/blob/master/Documentation/ABI/testing/procfs-diskstats
func ProcDiskstats() (*types.DiskStats, error) {
	file, err := os.Open(PROC_DISKSTATS)
	if err != nil {
		return nil, fmt.Errorf("error: opening %s %v", PROC_DISKSTATS, err)
	}
	defer file.Close()

	diskStats := &types.DiskStats{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		// Skip partitions and loop devices, only aggregate physical disks
		deviceName := fields[2]
		if strings.Contains(deviceName, "loop") ||
			len(deviceName) > 3 && (deviceName[len(deviceName)-1] >= '0' && deviceName[len(deviceName)-1] <= '9') {
			continue
		}

		// Parse disk stats (cumulative for all physical disks)
		readsCompleted, _ := strconv.ParseUint(fields[3], 10, 64)
		readsMerged, _ := strconv.ParseUint(fields[4], 10, 64)
		sectorsRead, _ := strconv.ParseUint(fields[5], 10, 64)
		timeReading, _ := strconv.ParseUint(fields[6], 10, 64)
		writesCompleted, _ := strconv.ParseUint(fields[7], 10, 64)
		writesMerged, _ := strconv.ParseUint(fields[8], 10, 64)
		sectorsWritten, _ := strconv.ParseUint(fields[9], 10, 64)
		timeWriting, _ := strconv.ParseUint(fields[10], 10, 64)

		diskStats.ReadsCompleted += readsCompleted
		diskStats.ReadsMerged += readsMerged
		diskStats.SectorsRead += sectorsRead
		diskStats.TimeReading += timeReading
		diskStats.WritesCompleted += writesCompleted
		diskStats.WritesMerged += writesMerged
		diskStats.SectorsWritten += sectorsWritten
		diskStats.TimeWriting += timeWriting
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: reading %s: %v", PROC_DISKSTATS, err)
	}

	return diskStats, nil
}
