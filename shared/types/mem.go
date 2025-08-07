package types

import "fmt"

type MemoryStats struct {
	MemTotal     int `json:"mem_total"`
	MemFree      int `json:"mem_free"`
	MemAvailable int `json:"mem_available"`
	Buffers      int `json:"buffers"`
	Cached       int `json:"cached"`
	SwapCached   int `json:"swap_cached"`
	SwapTotal    int `json:"swap_total"`
	SwapFree     int `json:"swap_free"`
}

func (m *MemoryStats) FormatMemory() string {
	if m.MemTotal == 0 {
		return "0%"
	}
	totalGB := float64(m.MemTotal) / 1024 / 1024
	availableGB := float64(m.MemAvailable) / 1024 / 1024
	usedGB := totalGB - availableGB
	usedPct := (usedGB / totalGB) * 100
	return fmt.Sprintf("%.1f%%", usedPct)
}
