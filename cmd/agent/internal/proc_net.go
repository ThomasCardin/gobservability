package internal

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/ThomasCardin/gobservability/cmd/agent/shared"
	"github.com/ThomasCardin/gobservability/shared/types"
)

func getProcNetDev(devMode string) string {
	return shared.GetProcBasePath(devMode) + "/net/dev"
}

// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.rst#13-networking-info-in-procnet
func ProcNetDev(devMode string) (*types.NetworkStats, error) {
	procNetDevPath := getProcNetDev(devMode)
	file, err := os.Open(procNetDevPath)
	if err != nil {
		return nil, errors.New("failed to open proc net/dev")
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
		return nil, errors.New("failed to read proc net/dev")
	}

	return netStats, nil
}
