package tflops

import (
	gpu "SuperNet-Node/machine_info/gpu/gpu_infos"
	logs "SuperNet-Node/utils/log_utils"
	"strings"
)

type InfoTFLOPS struct {
	TFLOPS float32 `json:"TFLOPS"`
}

func GetFlopsInfo(gpuName string) InfoTFLOPS {
	logs.Normal("Getting TFLOPS info...")

	gpuInfos := gpu.InitGpuInfos()
	for _, info := range gpuInfos {
		if strings.Contains(info.Name, gpuName) {
			return InfoTFLOPS{TFLOPS: info.Fp32}
		}
	}
	return InfoTFLOPS{TFLOPS: 0}
}
