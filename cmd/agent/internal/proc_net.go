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
	PROC_NET_DEV = "/proc/net/dev"
)

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#13-networking-info-in-procnet
func ProcNetDev() (*types.NetworkStats, error) {
	file, err := os.Open(PROC_NET_DEV)
	if err != nil {
		return nil, fmt.Errorf("error: opening %s %v", PROC_NET_DEV, err)
	}
	defer file.Close()

	netStats := &types.NetworkStats{}
	scanner := bufio.NewScanner(file)

	// Skip first two header lines
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 17 {
			continue
		}

		// Skip loopback interface
		interfaceName := strings.TrimSuffix(fields[0], ":")
		if interfaceName == "lo" {
			continue
		}

		// Parse network stats (cumulative for all non-loopback interfaces)
		rxBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[2], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[3], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[9], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[10], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[11], 10, 64)

		netStats.BytesReceived += rxBytes
		netStats.PacketsReceived += rxPackets
		netStats.ErrorsReceived += rxErrors
		netStats.BytesTransmitted += txBytes
		netStats.PacketsTransmitted += txPackets
		netStats.ErrorsTransmitted += txErrors
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: reading %s: %v", PROC_NET_DEV, err)
	}

	return netStats, nil
}
