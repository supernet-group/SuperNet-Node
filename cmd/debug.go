package cmd

import (

	// "github.com/jdgcs/ed25519/extra25519"

	// "github.com/docker/docker/api/types/container"
	// "github.com/docker/docker/client"
	// "github.com/docker/go-connections/nat"

	"SuperNet-Node/pattern"
	logs "SuperNet-Node/utils/log_utils"

	"github.com/urfave/cli"
	// "golang.org/x/crypto/nacl/box"
)

var DebugCommand = cli.Command{
	Name:  "debug",
	Usage: "debug.",
	Action: func(c *cli.Context) error {

		/* Debug : Force Complete */
		// machineUUID, err := machine_uuid.GetInfoMachineUUID()
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

		// metadata := "{\"formData\":{\"duration\":2,\"taskName\":\"Computing Task-48\"},\"MachineInfo\":{\"Uuid\":\"0xec29d2a3323b6812b091d70610d2e17b\",\"Provider\":\"AxBoDKGYKBa54qkDusWWYgf8QXufvBKTJTQBaKyEiEzF\",\"Region\":\"United States\",\"GPU\":\"1xNVIDIA A10G\",\"CPU\":\"AMD EPYC 7R32\",\"Tflops\":31.52,\"RAM\":\"15GB\",\"AvailDiskStorage\":100,\"Reliability\":\"100%\",\"CPS\":\"53.15\",\"Speed\":{\"Upload\":\"1180.81 Mbit/s\",\"Download\":\"1980.73 Mbit/s\"},\"MaxDuration\":10,\"Price\":1,\"IP\":\"44.220.54.244\",\"Port\":\"80\"},\"OrderInfo\":{\"Model\":22,\"Intent\":\"deploy\",\"DownloadURL\":[\"QmQ54jKv8A3XjcD2qs8HAyuyFotdkzXfZ2baMk6Fv9a3fy\"]}}"

		// err = json.Unmarshal([]byte(metadata), &orderPlacedMetadata)
		// if err != nil {
		// 	return err
		// }

		// orderPlacedMetadata.MachineAccounts = chainInfo.ProgramSuperMachine.String()

		// chainInfo.ProgramSuperOrder = solana.MustPublicKeyFromBase58("7cSAPAPqm2tKA5yyKAz8vxvKVu51oTSDXSVnpsPVr9YV")

		// buyer := solana.MustPublicKeyFromBase58("AxBoDKGYKBa54qkDusWWYgf8QXufvBKTJTQBaKyEiEzF")

		// superWrapper := super.NewSuperWrapper(chainInfo)
		// _, err = superWrapper.OrderFailed(buyer, orderPlacedMetadata)
		// if err != nil {
		// 	logs.Error(err.Error())
		// }

		// db, err := dbutils.NewDB()
		// if err != nil {
		// 	logs.Error(err.Error())
		// }
		// db.Delete([]byte("buyer"))
		// db.Delete([]byte("token"))
		// db.Close()
		// nginx.StopNginx()

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

		logs.Normal(pattern.LOGO)
		return nil
	},
}
