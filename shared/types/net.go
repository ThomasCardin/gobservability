package types

import "fmt"

type NetworkStats struct {
	BytesReceived      uint64 `json:"bytes_received"`
	BytesTransmitted   uint64 `json:"bytes_transmitted"`
	PacketsReceived    uint64 `json:"packets_received"`
	PacketsTransmitted uint64 `json:"packets_transmitted"`
	ErrorsReceived     uint64 `json:"errors_received"`
	ErrorsTransmitted  uint64 `json:"errors_transmitted"`
}

func (n *NetworkStats) FormatNetwork() string {
	totalMB := float64(n.BytesReceived+n.BytesTransmitted) / 1024 / 1024
	return fmt.Sprintf("%.1fM", totalMB)
}
