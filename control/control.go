package control

import (
	"SuperNet-Node/chain"
	"SuperNet-Node/chain/super"
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
	"time"

	"github.com/gagliardetto/solana-go"
)

// OrderComplete marks the completion of an order process.
func OrderComplete(super *super.WrapperSuper, metadata string, isGPU bool, containerID string) error {
	logs.Normal("Order is complete")
	// Stop the workspace container associated with the order.
	if err := docker.StopWorkspaceContainer(containerID); err != nil {
		return err
	}
	// Unmarshal the order placement metadata JSON string into a structured object.
	var orderPlacedMetadata pattern.OrderPlacedMetadata

	err := json.Unmarshal([]byte(metadata), &orderPlacedMetadata)
	if err != nil {
		return err
	}
	orderPlacedMetadata.MachineAccounts = super.ProgramSuperMachine.String()
	_, err = super.OrderCompleted(orderPlacedMetadata, isGPU)
	if err != nil {
		return err
	}
	return nil
}

// OrderFailed Handles the scenario where an order has failed.
func OrderFailed(super *super.WrapperSuper, orderPlacedMetadata pattern.OrderPlacedMetadata, buyer solana.PublicKey) error {
	logs.Normal("Order is failed")
	orderPlacedMetadata.MachineAccounts = super.ProgramSuperMachine.String()
	_, err := super.OrderFailed(buyer, orderPlacedMetadata)
	if err != nil {
		// Return a formatted error if the order fail processing encounters an issue
		return fmt.Errorf("> super.OrderFailed: %v", err.Error())
	}
	return nil
}

func GetSuper(longTime bool) (*super.WrapperSuper, *machine_info.MachineInfo, error) {

	var hwInfo machine_info.MachineInfo

	// Retrieve basic machine information; if an error occurs, return with an error message.
	hwInfo, err := machine_info.GetMachineInfo(longTime)
	if err != nil {
		return nil, nil, fmt.Errorf("> GetMachineInfo: %v", err)
	}

	// Gather disk information and attach it to the hardware info.
	diskInfo, err := disk.GetDiskInfo()
	if err != nil {
		return nil, nil, err
	}
	hwInfo.DiskInfo = diskInfo

	// Perform extended operations when longTime is true for detailed debugging and scoring.
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

	// Derive a unique machine UUID considering various hardware specifics; return an error if unable to do so.
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

	// Initialize a new configuration and fetch chain information using the machine UUID.
	newConfig := config.NewConfig(
		key,
		config.GlobalConfig.Base.Rpc)

	var chainInfo *chain.InfoChain
	chainInfo, err = chain.GetChainInfo(newConfig, machineUUID)
	if err != nil {
		return nil, nil, fmt.Errorf("> GetChainInfo: %v", err)
	}

	// Update hardware info with chain details and UUID.
	hwInfo.MachineAccounts = chainInfo.ProgramSuperMachine.String()
	hwInfo.Addr = chainInfo.Wallet.Wallet.PublicKey().String()
	hwInfo.MachineUUID = machineUUID

	// Log the hardware information in a human-readable format.
	jsonData, _ := json.Marshal(hwInfo)
	logs.Normal(fmt.Sprintf("Hardware Info : %v", string(jsonData)))

	// Ensure the directory for model creation exists; handle any errors during creation.
	modleCreatePath := pattern.ModleCreatePath
	err = os.MkdirAll(modleCreatePath, 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("> MkdirAll: %v", err)
	}
	// Define and download model-related static resources, unzip them, and log the completion.
	var modelURL []utils.DownloadURL
	modelURL = append(modelURL, utils.DownloadURL{
		URL:      config.GlobalConfig.Console.IpfsNodeUrl + "/ipfs" + utils.EnsureLeadingSlash("QmZ4eLbxWayopfecKTUxuAwYxjFg8yWKTrFB9z5CN82B6n"),
		Checksum: "",
		Name:     "DistriAI-Model-Create.zip",
	})
	err = utils.DownloadFiles(modleCreatePath, modelURL)
	if err != nil {
		return nil, nil, fmt.Errorf("> DownloadFiles: %v", err)
	}
	_, err = utils.Unzip(modleCreatePath+"/DistriAI-Model-Create.zip", modleCreatePath)
	if err != nil {
		return nil, nil, fmt.Errorf("> Unzip: %v", err)
	}
	logs.Normal("Model upload web static resources have been downloaded")
	return super.NewSuperWrapper(chainInfo), &hwInfo, nil
}

// var oldDuration time.Time
// var orderTimer *time.Timer

// OrderRefunded Handles the logic for processing a refunded order.
func OrderRefunded(containerID string) error {
	logs.Normal("Order is refunded")
	if err := docker.StopWorkspaceContainer(containerID); err != nil {
		return err
	}
	return nil
}

// StartHeartbeatTask starts a ticker to periodically submit heartbeat tasks.
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
