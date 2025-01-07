package cmd

import (
	"SuperNet-Node/config"
	"SuperNet-Node/control"
	"SuperNet-Node/docker"
	"SuperNet-Node/nginx"
	"SuperNet-Node/pattern"
	"SuperNet-Node/server"
	"SuperNet-Node/utils"
	dbutils "SuperNet-Node/utils/db_utils"
	logs "SuperNet-Node/utils/log_utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/urfave/cli"
)

var ClientCommand = cli.Command{
	Name:  "node",
	Usage: "Starting or terminating a node program.",
	Subcommands: []cli.Command{
		{
			Name:  "start",
			Usage: "Upload hardware configuration and initiate listening events.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "preload, L",
					Value: "n",
					Usage: "Preload AI models during idle time at night.",
				},
			},
			Action: func(c *cli.Context) error {
				defer dbutils.CloseDB()

				preload := c.String("preload")

				superWrapper, hwInfo, err := control.GetSuper(true)
				if err != nil {
					logs.Error(fmt.Sprintf("GetSuper: %v", err))
					return nil
				}

				// Easy debugging
				if err = nginx.StartNginx(
					config.GlobalConfig.Console.SuperPort,
					config.GlobalConfig.Console.WorkPort,
					config.GlobalConfig.Console.ServerPort); err != nil {
					logs.Error(fmt.Sprintf("StartNginx error: %v", err))
					return nil
				}

				machine, err := superWrapper.GetMachine()
				if err != nil {
					logs.Error(fmt.Sprintf("GetMachine: %v", err))
					return nil
				}

				if machine.Metadata == "" {
					logs.Normal("Machine does not exist")
					_, err := superWrapper.AddMachine(*hwInfo)
					if err != nil {
						logs.Error(fmt.Sprintf("AddMachine: %v", err))
						return nil
					}
				} else {
					logs.Normal("Machine already exists")
				}

				logs.Normal(pattern.LOGO)

				go server.StartServer(config.GlobalConfig.Console.ServerPort)

				control.StartHeartbeatTask(superWrapper, hwInfo.MachineUUID)

				for {
					time.Sleep(1 * time.Minute)

					machine, err = superWrapper.GetMachine()
					if err != nil {
						logs.Error(fmt.Sprintf("GetMachine: %v", err))
						continue
					}

				ListenLoop:
					switch machine.Status.String() {
					case "Idle":
						// TODO: Add the logic of the Idle status.
						break ListenLoop
					case "ForRent":
						if preload != "y" && preload != "yes" {
							break ListenLoop
						}
						// Easy debug
						if !utils.IsLateNight() {
							break ListenLoop
						}

						if utils.RandomInt(26) != 1 {
							break ListenLoop
						}
						// if utils.RandomInt(3) != 1 {
						// 	logs.Normal("Not in the preload range")
						// 	break ListenLoop
						// }

						err = control.IdlePreload(superWrapper.Wallet.Wallet.PublicKey().String(), string(hwInfo.MachineUUID), hwInfo.DiskInfo.TotalSpace)
						if err != nil {
							logs.Error(fmt.Sprintf("IdlePreload: %v", err))
						}
						break ListenLoop
					case "Renting":

						logs.Normal(fmt.Sprintf("Machine is Renting, Details: %v", machine))

						orderID := machine.OrderPda
						if orderID.Equals(solana.SystemProgramID) {
							logs.Error(fmt.Sprintf("machine OrderPda error, OrderPda: %v", orderID))
							break ListenLoop
						}

						superWrapper.ProgramSuperOrder = orderID
						newOrder, err := superWrapper.GetOrder()
						if err != nil {
							logs.Error(fmt.Sprintf("GetOrder Error: %v", err))
							break ListenLoop
						}

						var orderPlacedMetadata pattern.OrderPlacedMetadata

						err = json.Unmarshal([]byte(newOrder.Metadata), &orderPlacedMetadata)
						if err != nil {
							logs.Error(fmt.Sprintf("json.Unmarshal: %v", err))
							break ListenLoop
						}

						isGPU := false
						if hwInfo.GPUInfo.Number > 0 {
							isGPU = true
						}

						var containerID string

						switch orderPlacedMetadata.OrderInfo.Intent {
						case "train":
							mlToken, err := dbutils.GenToken(newOrder.Buyer.String())
							if err != nil {
								logs.Error(fmt.Sprintf("GenToken: %v", err))
								break ListenLoop
							}
							logs.Normal(fmt.Sprintf("From buyer: %v ; mlToken: %v", newOrder.Buyer, mlToken))

							containerID, err = docker.TestRunWorkspaceContainer(isGPU, mlToken)
							if err != nil {
								logs.Error(fmt.Sprintln("RunWorkspaceContainer error: ", err))
								orderPlacedMetadata.OrderInfo.Message = err.Error()
								if err = control.OrderFailed(superWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
									logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
								}
								break ListenLoop
							}

							modelNames := orderPlacedMetadata.OrderInfo.DownloadURL

							if len(modelNames) <= 0 {
								break ListenLoop
							}
							if len(modelNames) > 5 {
								errMsg := "The number of models exceeds the limit"
								logs.Error(errMsg)
								orderPlacedMetadata.OrderInfo.Message = errMsg
								if err = control.OrderFailed(superWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
									logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
								}
								break ListenLoop
							}

							fileNames, err := utils.ListFiles(config.GlobalConfig.Console.WorkDirectory + "/" + pattern.IdlePreload)
							if err != nil {
								logs.Error(fmt.Sprintf("ListFiles: %v", err))
								break ListenLoop
							}
							fileNames = utils.FilterStrings(fileNames, ".zip")
							logs.Normal(fmt.Sprintf("Current local AI model: %v", fileNames))

							for _, modelName := range modelNames {
								isExist := false
								for _, fileName := range fileNames {
									if strings.Contains(fileName, modelName) {
										isExist = true
										break
									}
								}

								if !isExist {
									errMsg := fmt.Sprintf("Model file does not exist, URL: %v", modelName)
									logs.Error(errMsg)
									orderPlacedMetadata.OrderInfo.Message = errMsg
									if err = control.OrderFailed(superWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
										logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
									}
									break ListenLoop
								}
							}

							modelDir := config.GlobalConfig.Console.WorkDirectory + "/ml-workspace"

							for _, modelName := range modelNames {
								name := modelName
								idx := strings.Index(modelName, "-")
								if idx != -1 {
									name = modelName[idx+1:]
								}

								_, err = utils.Unzip(
									config.GlobalConfig.Console.WorkDirectory+"/"+pattern.IdlePreload+"/"+modelName+".zip",
									modelDir+"/"+name)
								if err != nil {
									logs.Error(fmt.Sprintf("Unzip: %v", err))
									break ListenLoop
								}
							}
						case "deploy":
							_, err := dbutils.GenToken(newOrder.Buyer.String())
							if err != nil {
								logs.Error(fmt.Sprintf("GenToken: %v", err))
								break ListenLoop
							}

							var downloadDeployURL []string

							url := orderPlacedMetadata.OrderInfo.DownloadURL
							if len(url) > 0 {
								deployDir := config.GlobalConfig.Console.WorkDirectory
								var deployURL []utils.DownloadURL
								deployURL = append(deployURL, utils.DownloadURL{
									URL:      config.GlobalConfig.Console.IpfsNodeUrl + "/ipfs" + utils.EnsureLeadingSlash(url[0]),
									Checksum: "",
									Name:     "CID.json",
								})

								logs.Normal("Downloading CID.json ...")
								err = utils.DownloadFiles(deployDir, deployURL)
								if err != nil {
									logs.Error(fmt.Sprintf("DownloadFiles: %v", err))
								}

								items, err := utils.GetCidItemsFromFile(deployDir + "/CID.json")
								if err != nil {
									logs.Error(fmt.Sprintf("GetCidItemsFromFile: %v", err))
								}

								err = os.Remove(deployDir + "/CID.json")
								if err != nil {
									logs.Error(fmt.Sprintf("Remove CID.json: %v", err))
								}

								for _, item := range items {
									downloadDeployURL = append(downloadDeployURL, config.GlobalConfig.Console.IpfsNodeUrl+utils.EnsureLeadingSlash(item.Cid))
								}
							}

							logs.Normal("Run deploy container ...")
							logs.Normal(fmt.Sprintf("DownloadDeployURL: %v", downloadDeployURL))

							containerID, err = docker.RunDeployContainer(isGPU, downloadDeployURL)
							if err != nil {
								logs.Error(fmt.Sprintln("RunDeployContainer error ", err))
								orderPlacedMetadata.OrderInfo.Message = err.Error()
								if err = control.OrderFailed(superWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
									logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
								}
								break ListenLoop
							}
						default:
							logs.Error(fmt.Sprintf("OrderInfo.Intent error, Intent: %v", orderPlacedMetadata.OrderInfo.Intent))
							break ListenLoop
						}

						_, err = superWrapper.OrderStart()
						if err != nil {
							logs.Error(fmt.Sprintf("OrderStart: %v", err))
							if err := docker.StopWorkspaceContainer(containerID); err != nil {
								logs.Error(fmt.Sprintf("> StopWorkspaceContainer, containerID: %s, err: %v", containerID, err))
							}
							break ListenLoop
						}

						for {
							time.Sleep(1 * time.Minute)

							newOrder, err = superWrapper.GetOrder()
							if err != nil {
								logs.Error(fmt.Sprintf("GetOrder Error: %v", err))
								break ListenLoop
							}

							switch newOrder.Status.String() {
							case "Preparing":
								logs.Error(fmt.Sprintf("Order error, ID: %v\norder: %v", superWrapper.ProgramSuperOrder, newOrder))
								break ListenLoop
							case "Training":
								orderEndTime := time.Unix(newOrder.StartTime, 0).Add(time.Hour * time.Duration(newOrder.Duration))

								db := dbutils.GetDB()
								dbutils.Update(db, []byte("orderEndTime"), []byte(orderEndTime.Format(time.RFC3339)))

								timeNow := time.Now()
								if timeNow.After(orderEndTime) {

									logs.Normal(fmt.Sprintf("Order completed, Details: %v", newOrder))

									if err = control.OrderComplete(superWrapper, newOrder, isGPU, containerID); err != nil {
										logs.Error(fmt.Sprintf("OrderComplete: %v", err))
									}
									break ListenLoop
								}
								continue
							case "Completed":
								logs.Error(fmt.Sprintf("Order error, ID: %v\norder: %v", superWrapper.ProgramSuperOrder, newOrder))
								break ListenLoop
							case "Failed":
								logs.Error(fmt.Sprintf("Order error, ID: %v\norder: %v", superWrapper.ProgramSuperOrder, newOrder))
								break ListenLoop
							case "Refunded":
								err = control.OrderRefunded(containerID)
								if err != nil {
									logs.Error(fmt.Sprintf("OrderRefunded: %v", err))
								}
								break ListenLoop
							}
						}
					default:
						logs.Error(fmt.Sprintf("machine status error, Status: %v", machine.Status))
						break ListenLoop
					}
				}
			},
		},
		{
			Name:  "stop",
			Usage: "Stop the client.",
			Action: func(c *cli.Context) error {
				nginx.StopNginx()

				superWrapper, _, err := control.GetSuper(false)
				if err != nil {
					logs.Error(err.Error())
					return nil
				}

				hash, err := superWrapper.RemoveMachine()
				if err != nil {
					logs.Error(fmt.Sprintf("Error block : %v, msg : %v\n", hash, err))
				}

				db := dbutils.GetDB()
				defer dbutils.CloseDB()
				dbutils.Delete(db, []byte("buyer"))
				dbutils.Delete(db, []byte("token"))
				dbutils.Delete(db, []byte("orderEndTime"))
				dbutils.CloseDB()

				err = os.RemoveAll(pattern.ModleCreatePath)
				if err != nil {
					logs.Error(fmt.Sprintf("RemoveAll: %v", err))
				}

				return nil
			},
		},
	},
}
