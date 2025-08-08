package types

import "fmt"

type NetworkStats struct {
	// Raw values from /proc/net/dev
	BytesReceived      uint64 `json:"bytes_received"`
	BytesTransmitted   uint64 `json:"bytes_transmitted"`
	PacketsReceived    uint64 `json:"packets_received"`
	PacketsTransmitted uint64 `json:"packets_transmitted"`
	ErrorsReceived     uint64 `json:"errors_received"`
	ErrorsTransmitted  uint64 `json:"errors_transmitted"`
	
	// Calculated rates by agent (MB/s)
	RxRate    float64 `json:"rx_rate"`    // Receive rate in MB/s
	TxRate    float64 `json:"tx_rate"`    // Transmit rate in MB/s
	TotalRate float64 `json:"total_rate"` // Total rate in MB/s
}

func (n *NetworkStats) FormatNetwork() string {
	totalMB := float64(n.BytesReceived+n.BytesTransmitted) / 1024 / 1024
	return fmt.Sprintf("%.1fM", totalMB)
}
