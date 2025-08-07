package types

import "fmt"

type DiskStats struct {
	ReadsCompleted  uint64 `json:"reads_completed"`
	ReadsMerged     uint64 `json:"reads_merged"`
	SectorsRead     uint64 `json:"sectors_read"`
	TimeReading     uint64 `json:"time_reading"`
	WritesCompleted uint64 `json:"writes_completed"`
	WritesMerged    uint64 `json:"writes_merged"`
	SectorsWritten  uint64 `json:"sectors_written"`
	TimeWriting     uint64 `json:"time_writing"`
}

func (d *DiskStats) FormatDisk() string {
	totalMB := float64(d.SectorsRead+d.SectorsWritten) * 512 / 1024 / 1024
	return fmt.Sprintf("%.1fM", totalMB)
}
