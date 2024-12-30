package disk

import (
	"SuperNet-Node/config"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"fmt"
	"os"
)

type InfoDisk struct {
	Path       string  `json:"Path"`
	TotalSpace float64 `json:"TotalSpace"`
	// FreeSpace  float64 `json:"FreeSpace"`
}

func GetDiskInfo() (InfoDisk, error) {
	logs.Normal("Getting free space info...")

	dirpath := config.GlobalConfig.Console.WorkDirectory
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		logs.Normal(fmt.Sprintf("%s does not exist. Using default directory: /data/distri", dirpath))
		dirpath = "/data/distri"
		os.MkdirAll(dirpath, 0755)
		config.GlobalConfig.Console.WorkDirectory = dirpath
	}

	freeSpace, err := utils.GetFreeSpace(dirpath)
	if err != nil {
		return InfoDisk{}, fmt.Errorf("error calculating free space: %v", err)
	}

	diskInfo := InfoDisk{
		Path:       dirpath,
		TotalSpace: float64(freeSpace) / 1024 / 1024 / 1024,
	}
	return diskInfo, nil
}
