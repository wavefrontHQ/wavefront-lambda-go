package wflambda

import "github.com/shirou/gopsutil/mem"

type memStats struct {
	Total          float64
	Used           float64
	UsedPercentage float64
}

// getMemoryStats gets memory statistics from the container the AWS Lambda function runs in
func getMemoryStats() *memStats {
	stats, _ := mem.VirtualMemory()
	return &memStats{
		Total:          float64(stats.Total) / float64(1<<20),
		Used:           float64(stats.Used) / float64(1<<20),
		UsedPercentage: stats.UsedPercent,
	}
}
