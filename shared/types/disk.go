package types

import "fmt"

type DiskStats struct {
	// Raw values from /proc/diskstats
	ReadsCompleted  uint64 `json:"reads_completed"`
	ReadsMerged     uint64 `json:"reads_merged"`
	SectorsRead     uint64 `json:"sectors_read"`
	TimeReading     uint64 `json:"time_reading"`
	WritesCompleted uint64 `json:"writes_completed"`
	WritesMerged    uint64 `json:"writes_merged"`
	SectorsWritten  uint64 `json:"sectors_written"`
	TimeWriting     uint64 `json:"time_writing"`

	// Calculated rates by agent (MB/s)
	ReadRate  float64 `json:"read_rate"`  // Read rate in MB/s
	WriteRate float64 `json:"write_rate"` // Write rate in MB/s
	TotalRate float64 `json:"total_rate"` // Total I/O rate in MB/s
}

func (d *DiskStats) FormatDisk() string {
	totalMB := float64(d.SectorsRead+d.SectorsWritten) * 512 / 1024 / 1024
	return fmt.Sprintf("%.1fM", totalMB)
}
