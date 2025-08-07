package types

import "fmt"

type CPUStats struct {
	User    int `json:"user"`
	Nice    int `json:"nice"`
	System  int `json:"system"`
	Idle    int `json:"idle"`
	IOWait  int `json:"iowait"`
	IRQ     int `json:"irq"`
	SoftIRQ int `json:"softirq"`
	Steal   int `json:"steal"`
	Total   int `json:"total"`
}

func (c *CPUStats) FormatCPU() string {
	if c.Total == 0 {
		return "0%"
	}
	total := float64(c.Total)
	userTotal := float64(c.User + c.Nice)
	systemTotal := float64(c.IRQ + c.SoftIRQ)
	activePct := (userTotal + systemTotal + float64(c.System)) * 100 / total
	return fmt.Sprintf("%.1f%%", activePct)
}
