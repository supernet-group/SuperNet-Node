package control

import (
	"SuperNet-Node/api"
	"SuperNet-Node/chain"
	"SuperNet-Node/chain/super"
	"SuperNet-Node/chain/super/supernet"
	"SuperNet-Node/config"
	"SuperNet-Node/docker"
	"SuperNet-Node/machine_info"
	"SuperNet-Node/machine_info/disk"
	"SuperNet-Node/machine_info/machine_uuid"
	"SuperNet-Node/pattern"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
)

func OrderComplete(super *super.WrapperSuper, order supernet.Order, isGPU bool, containerID string) error {
	logs.Normal("Order is complete")

	if err := docker.StopWorkspaceContainer(containerID); err != nil {
		return err
	}

	_, err := super.OrderCompleted(order, isGPU)
	if err != nil {
		return err
	}
	return nil
}

func OrderFailed(super *super.WrapperSuper, orderPlacedMetadata pattern.OrderPlacedMetadata, buyer solana.PublicKey) error {
	logs.Normal("Order is failed")

	_, err := super.OrderFailed(buyer, orderPlacedMetadata)
	if err != nil {
		return fmt.Errorf("> super.OrderFailed: %v", err.Error())
	}
	return nil
}

func GetSuper(longTime bool) (*super.WrapperSuper, *machine_info.MachineInfo, error) {

	var hwInfo machine_info.MachineInfo

	hwInfo, err := machine_info.GetMachineInfo(longTime)
	if err != nil {
		return nil, nil, fmt.Errorf("> GetMachineInfo: %v", err)
	}

	diskInfo, err := disk.GetDiskInfo()
	if err != nil {
		return nil, nil, err
	}
	hwInfo.DiskInfo = diskInfo

	if longTime {
		// Easy debugging
		isGPU := false
		if hwInfo.GPUInfo.Number > 0 {
			isGPU = true
		}
		score, err := docker.RunScoreContainer(isGPU)
		if err != nil {
			return nil, nil, err
		}

		imageWorkspace := pattern.ML_WORKSPACE_NAME
		if isGPU {
			imageWorkspace = pattern.ML_WORKSPACE_GPU_NAME
		}
		if err = docker.ImageExistOrPull(imageWorkspace); err != nil {
			return nil, nil, err
		}

		hwInfo.Score = score
	}

	key := config.GlobalConfig.Base.PrivateKey

	machineUUID, err := machine_uuid.GetInfoMachineUUID(
		hwInfo.CPUInfo.ModelName,
		hwInfo.GPUInfo.Model,
		hwInfo.IpInfo.IP,
		hwInfo.LocationInfo.Country,
		hwInfo.LocationInfo.Region,
		hwInfo.LocationInfo.City)
	if err != nil {
		return nil, nil, fmt.Errorf("> GetInfoMachineUUID: %v", err)
	}

	newConfig := config.NewConfig(
		key,
		config.GlobalConfig.Base.Rpc)

	var chainInfo *chain.InfoChain
	chainInfo, err = chain.GetChainInfo(newConfig, machineUUID)
	if err != nil {
		return nil, nil, fmt.Errorf("> GetChainInfo: %v", err)
	}

	hwInfo.MachineAccounts = chainInfo.ProgramSuperMachine.String()
	hwInfo.Addr = chainInfo.Wallet.Wallet.PublicKey().String()
	hwInfo.MachineUUID = machineUUID

	jsonData, _ := json.Marshal(hwInfo)
	logs.Normal(fmt.Sprintf("Hardware Info : %v", string(jsonData)))

	// Easy debugging
	modleCreatePath := pattern.ModleCreatePath
	err = os.MkdirAll(modleCreatePath, 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("> MkdirAll: %v", err)
	}
	var modelURL []utils.DownloadURL
	modelURL = append(modelURL, utils.DownloadURL{
		URL:      config.GlobalConfig.Console.IpfsNodeUrl + "/ipfs" + utils.EnsureLeadingSlash(pattern.ModelCreateCID),
		Checksum: "",
		Name:     pattern.ModelCreateName + ".zip",
	})
	err = utils.DownloadFiles(modleCreatePath, modelURL)
	if err != nil {
		return nil, nil, fmt.Errorf("> DownloadFiles: %v", err)
	}
	_, err = utils.Unzip(modleCreatePath+"/"+pattern.ModelCreateName+".zip", modleCreatePath)
	if err != nil {
		return nil, nil, fmt.Errorf("> Unzip: %v", err)
	}
	logs.Normal("Model upload web static resources have been downloaded")

	return super.NewSuperWrapper(chainInfo), &hwInfo, nil
}

func OrderRefunded(containerID string) error {
	logs.Normal("Order is refunded")
	if err := docker.StopWorkspaceContainer(containerID); err != nil {
		return err
	}
	return nil
}

// temp
func StartHeartbeatTask(super *super.WrapperSuper, machineID machine_uuid.MachineUUID) {
	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				taskID, err := utils.GenerateRandomString(16)
				if err != nil {
					logs.Error(err.Error())
				}
				taskUuid, err := utils.ParseTaskUUID(string(taskID))
				if err != nil {
					logs.Error(fmt.Sprintf("error parsing taskUuid: %v", err))
				}

				machineUuid, err := utils.ParseMachineUUID(string(machineID))
				if err != nil {
					logs.Error(fmt.Sprintf("error parsing machineUuid: %v", err))
				}

				hash, err := super.SubmitTask(taskUuid, machineUuid, utils.CurrentPeriod(), pattern.TaskMetadata{})
				if err != nil {
					logs.Error(fmt.Sprintf("Error block : %v, msg : %v\n", hash, err))
				}
			}
		}
	}()
}

func IdlePreload(ownerAddr string, machineUUID string, totalDiskSpace float64) error {
	utils.EnsureDirExists(config.GlobalConfig.Console.WorkDirectory + "/" + pattern.IdlePreload)

	resModelList, err := api.GetModelList()
	if err != nil {
		return err
	}

	if len(resModelList.Data.List) <= 0 {
		logs.Warning("No model available")
		return nil
	}

	fileNames, err := utils.ListFiles(config.GlobalConfig.Console.WorkDirectory + "/" + pattern.IdlePreload)
	if err != nil {
		return err
	}
	fileNames = utils.FilterStrings(fileNames, ".zip")

	obsolete := resModelList.IsObsolete(fileNames)

	var cachedModels []api.Cached
	var cachedDatasets []api.Cached

	if len(obsolete) != 0 {
		for _, newFileName := range utils.DiffStrings(fileNames, obsolete) {
			idx := strings.Index(newFileName, "-")
			if idx != -1 {
				cachedModels = append(cachedModels, api.Cached{
					Owner: newFileName[:idx],
					Name:  newFileName[idx+1:],
				})
			}
		}

		_, err = api.UpdateModelCached(ownerAddr, machineUUID, cachedModels, cachedDatasets)
		if err != nil {
			return err
		}

		for _, fileName := range obsolete {
			err = os.Remove(config.GlobalConfig.Console.WorkDirectory + "/" + pattern.IdlePreload + "/" + fileName)
			if err != nil {
				return err
			}
		}
	} else {
		for _, fileName := range fileNames {
			idx := strings.Index(fileName, "-")
			if idx != -1 {
				cachedModels = append(cachedModels, api.Cached{
					Owner: fileName[:idx],
					Name:  fileName[idx+1:],
				})
			}
		}
	}

	if len(cachedModels) >= 5 {
		return nil
	}

	// resMode := resModelList.Data.List[utils.RandomInt(9)]
	resMode := resModelList.Data.List[utils.RandomInt(len(resModelList.Data.List)-1)]
	owner := resMode.Owner
	name := resMode.Name
	for _, model := range cachedModels {
		if model.Owner == owner && model.Name == name {
			return nil
		}
	}

	resFileStat, err := api.FileStatInIPFS(config.GlobalConfig.Console.IpfsNodeUrl, "/supernet/model/"+owner+"/"+name)
	if err != nil {
		return fmt.Errorf("> FileStatInIPFS: %v", err)
	}
	diskInfo, err := disk.GetDiskInfo()
	if err != nil {
		return fmt.Errorf("> GetDiskInfo: %v", err)
	}
	downloadableSpace := diskInfo.TotalSpace - (resFileStat.CumulativeSize / 1024 / 1024 / 1024)
	if downloadableSpace <= 100 {
		return nil
	}

	modelDir := config.GlobalConfig.Console.WorkDirectory + "/" + pattern.IdlePreload + "/" + fmt.Sprintf("%s-%s", owner, name)
	utils.EnsureDirExists(modelDir)

	logs.Normal(fmt.Sprintf("Preloading modelOwner : %s , modelName : %s", owner, name))

	err = api.DownloadProjectInIPFS(
		config.GlobalConfig.Console.IpfsNodeUrl,
		modelDir,
		resFileStat.Hash)
	if err != nil {
		return fmt.Errorf("> DownloadProjectInIPFS: %v", err)
	}

	err = utils.Zip(modelDir, modelDir+".zip")
	if err != nil {
		return fmt.Errorf("> Zip: %v", err)
	}

	cachedModels = append(cachedModels, api.Cached{
		Owner: owner,
		Name:  name,
	})

	_, err = api.UpdateModelCached(ownerAddr, machineUUID, cachedModels, cachedDatasets)
	if err != nil {
		return fmt.Errorf("> UpdateModelCached: %v", err)
	}
	return nil
}
