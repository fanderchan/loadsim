package system

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Stats struct {
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsedMB  uint64
	MemoryTotalMB uint64
}

func Snapshot(cpuSample time.Duration) (Stats, error) {
	if cpuSample <= 0 {
		cpuSample = 150 * time.Millisecond
	}

	percentages, err := cpu.Percent(cpuSample, false)
	if err != nil {
		return Stats{}, err
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return Stats{}, err
	}

	stats := Stats{
		MemoryPercent: memStat.UsedPercent,
		MemoryUsedMB:  memStat.Used / (1024 * 1024),
		MemoryTotalMB: memStat.Total / (1024 * 1024),
	}
	if len(percentages) > 0 {
		stats.CPUPercent = percentages[0]
	}

	return stats, nil
}
