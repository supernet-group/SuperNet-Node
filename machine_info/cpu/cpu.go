package cpu

import (
	"SuperNet-Node/utils/log_utils"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
)

type InfoCPU struct {
	ModelName string  `json:"ModelName"`
	Cores     int32   `json:"Cores"`
	Mhz       float64 `json:"Mhz"`
}

func GetCPUInfo() (InfoCPU, error) {
	logs.Normal("Getting CPU info...")

	// Easy debugging
	cpuInfoStats, err := cpu.Info()
	if err != nil {
		return InfoCPU{}, fmt.Errorf("failed to get CPU info: %w", err)
	}

	counts, err := cpu.Counts(false)
	if err != nil {
		return InfoCPU{}, fmt.Errorf("failed to get CPU Counts : %w", err)
	}

	cpuInfo := InfoCPU{
		ModelName: cpuInfoStats[0].ModelName,
		Cores:     int32(counts),
		Mhz:       cpuInfoStats[0].Mhz,
	}

	return cpuInfo, nil
}
