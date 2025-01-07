package machine_info

import (
	"SuperNet-Node/config"
	"SuperNet-Node/machine_info/cpu"
	"SuperNet-Node/machine_info/disk"
	"SuperNet-Node/machine_info/gpu"
	"SuperNet-Node/machine_info/ip"
	"SuperNet-Node/machine_info/location"
	"SuperNet-Node/machine_info/machine_uuid"
	"SuperNet-Node/machine_info/memory"
	"SuperNet-Node/machine_info/speedtest"
	"SuperNet-Node/machine_info/tflops"
	logs "SuperNet-Node/utils/log_utils"
	"fmt"
)

type MachineInfo struct {
	MachineUUID     machine_uuid.MachineUUID `json:"MachineUUID"`
	Addr            string                   `json:"Addr"`
	IpInfo          ip.InfoIP                `json:"Ip"`
	CPUInfo         cpu.InfoCPU              `json:"CPUInfo"`
	DiskInfo        disk.InfoDisk            `json:"DiskInfo"`
	Score           float64                  `json:"Score"`
	MemoryInfo      memory.InfoMemory        `json:"InfoMemory"`
	GPUInfo         gpu.InfoGPU              `json:"GPUInfo"`
	LocationInfo    location.InfoLocation    `json:"LocationInfo"`
	SpeedInfo       speedtest.InfoSpeed      `json:"SpeedInfo"`
	TFLOPSInfo      tflops.InfoTFLOPS        `json:"InfoTFLOPS"`
	SecurityLevel   string                   `json:"SecurityLevel"`
	MachineAccounts string                   `json:"MachineAccounts"`
}

func GetMachineInfo(longTime bool) (MachineInfo, error) {
	var hwInfo MachineInfo

	ipInfo, err := ip.GetIpInfo()
	if err != nil {
		return hwInfo, fmt.Errorf("> GetIpInfo: %v", err)
	}
	hwInfo.IpInfo = ipInfo

	locationInfo, err := location.GetLocationInfo(ipInfo.IP)
	if err != nil {
		return hwInfo, fmt.Errorf("> GetLocationInfo: %v", err)
	}
	hwInfo.LocationInfo = locationInfo

	cpuInfo, err := cpu.GetCPUInfo()
	if err != nil {
		return hwInfo, err
	}
	hwInfo.CPUInfo = cpuInfo

	memInfo, err := memory.GetMemoryInfo()
	if err != nil {
		return hwInfo, err
	}
	hwInfo.MemoryInfo = memInfo

	gpuInfo, _ := gpu.GetGPUInfo()
	// if err != nil {
	// 	return hwInfo, err
	// }
	hwInfo.GPUInfo = gpuInfo

	if longTime {
		// Easy debugging
		speedInfo, err := speedtest.GetSpeedInfo()
		if err != nil {
			logs.Warning(err.Error())
		}
		// speedInfo := speedtest.InfoSpeed{
		// 	Download: "61.90 Mbit/s",
		// 	Upload:   "59.49 Mbit/s",
		// }
		hwInfo.SpeedInfo = speedInfo
	}

	hwInfo.TFLOPSInfo = tflops.GetFlopsInfo(gpuInfo.Model)

	hwInfo.SecurityLevel = config.GlobalConfig.Base.SecurityLevel

	return hwInfo, nil
}
