package gpu

import (
	"SuperNet-Node/pattern"
	logs "SuperNet-Node/utils/log_utils"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type InfoGPU struct {
	Model  string `json:"Model"`
	Number int    `json:"Number"`
	Memory string `json:"Memory"`
}

// GetGPUInfo retrieves information about the GPU(s) installed on the system.
func GetGPUInfo() (InfoGPU, error) {
	logs.Normal("Getting GPU info...")

	var gpuInfo InfoGPU

	cmd := exec.Command("nvidia-smi", "--query-gpu=count,gpu_name,memory.total", "--format=csv,noheader")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// Easy debugging
		return InfoGPU{Model: pattern.NO_GPU, Number: 0}, err
		// return InfoGPU{Model: "NVIDIA Tesla V100 PCIe 16 GB", Number: 8}, err
	}

	result := strings.Split(strings.TrimSpace(out.String()), ", ")

	if len(result) >= 2 {

		number, err := strconv.Atoi(result[0])
		if err != nil {
			number = 0
		}

		gpuInfo = InfoGPU{
			Model:  result[1],
			Number: number,
			Memory: result[2],
		}
	}
	return gpuInfo, nil
}
