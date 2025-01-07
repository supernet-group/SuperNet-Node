package cmd

import (
	logs "SuperNet-Node/utils/log_utils"

	// "github.com/jdgcs/ed25519/extra25519"

	// "github.com/docker/docker/api/types/container"
	// "github.com/docker/docker/client"
	// "github.com/docker/go-connections/nat"

	"github.com/urfave/cli"
	// "golang.org/x/crypto/nacl/box"
)

var DebugCommand = cli.Command{
	Name:  "debug",
	Usage: "debug.",
	Action: func(c *cli.Context) error {
		logs.Normal("Debug")

		/* Debug : Force Complete */
		// nginx.StopNginx()

		// var hwInfo machine_info.MachineInfo

		// hwInfo, err := machine_info.GetMachineInfo(false)
		// if err != nil {
		// 	return err
		// }

		// diskInfo, err := disk.GetDiskInfo()
		// if err != nil {
		// 	return err
		// }
		// hwInfo.DiskInfo = diskInfo

		// machineUUID, err := machine_uuid.GetInfoMachineUUID(
		// 	hwInfo.CPUInfo.ModelName,
		// 	hwInfo.GPUInfo.Model,
		// 	hwInfo.IpInfo.IP,
		// 	hwInfo.LocationInfo.Country,
		// 	hwInfo.LocationInfo.Region,
		// 	hwInfo.LocationInfo.City)
		// if err != nil {
		// 	return err
		// }

		// key := config.GlobalConfig.Base.PrivateKey

		// newConfig := config.NewConfig(
		// 	key,
		// 	pattern.RPC)

		// var chainInfo *chain.InfoChain
		// chainInfo, err = chain.GetChainInfo(newConfig, machineUUID)
		// if err != nil {
		// 	return err
		// }

		// var orderPlacedMetadata pattern.OrderPlacedMetadata

		// metadata := "{\"formData\":{\"duration\":2,\"taskName\":\"Computing Task-28\"},\"MachineInfo\":{\"Uuid\":\"0x90899f35b7fa4c1e6cd51e3b03023a35\",\"Provider\":\"AxBoDKGYKBa54qkDusWWYgf8QXufvBKTJTQBaKyEiEzF\",\"Region\":\"United States\",\"GPU\":\"1x NVIDIA A10G\",\"CPU\":\"AMD EPYC 7R32\",\"Tflops\":31.52,\"RAM\":\"15\",\"AvailDiskStorage\":100,\"Reliability\":\"0%\",\"CPS\":\"52.77\",\"Speed\":{\"Upload\":\"496.68 Mbit/s\",\"Download\":\"698.15 Mbit/s\"},\"MaxDuration\":10,\"Price\":1,\"IP\":\"44.220.54.244\",\"Port\":\"80\"},\"OrderInfo\":{\"Model\":\"\",\"Intent\":\"train\",\"DownloadURL\":[]}}"

		// err = json.Unmarshal([]byte(metadata), &orderPlacedMetadata)
		// if err != nil {
		// 	return err
		// }

		// chainInfo.ProgramSuperOrder = solana.MustPublicKeyFromBase58("EdPB44J5h3WSw4xow1tez7cBsW6VZhFC9BmMjfA7QRoJ")

		// buyer := solana.MustPublicKeyFromBase58("AxBoDKGYKBa54qkDusWWYgf8QXufvBKTJTQBaKyEiEzF")

		// superWrapper := super.NewSuperWrapper(chainInfo)
		// // _, err = superWrapper.OrderFailed(buyer, orderPlacedMetadata)
		// // if err != nil {
		// // 	logs.Error(err.Error())
		// // }
		// _, err = superWrapper.OrderRefund(buyer)
		// if err != nil {
		// 	logs.Error(err.Error())
		// }

		// db := dbutils.GetDB()
		// defer dbutils.CloseDB()
		// dbutils.Delete(db, []byte("buyer"))
		// dbutils.Delete(db, []byte("token"))
		// dbutils.Delete(db, []byte("orderEndTime"))
		// dbutils.CloseDB()

		/* Debug : docker space */
		// sizeLimitGB := 10

		// result, err := utils.CompareSpaceWithDocker(sizeLimitGB)
		// if err != nil {
		// 	logs.Error(fmt.Sprintf("An error occurred while comparing disk space sizes: %v\n", err))
		// 	return nil
		// }

		// if result {
		// 	logs.Normal("There is sufficient free space in the Docker directory")
		// } else {
		// 	logs.Normal("Insufficient free space in the Docker directory")
		// }
		// return nil

		/* =============================================== */
		return nil
	},
}
