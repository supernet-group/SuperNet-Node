package memory

import (
	"SuperNet-Node/utils/log_utils"
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
)

type InfoMemory struct {
	RAM float64 `json:"RAM"`
}

func GetMemoryInfo() (InfoMemory, error) {
	logs.Normal("Getting memory info...")

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return InfoMemory{}, fmt.Errorf("failed to get memory info: %w", err)
	}

	return InfoMemory{
		RAM: float64(vmStat.Total) / (1 << 30),
	}, nil
}
