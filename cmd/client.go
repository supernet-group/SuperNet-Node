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

				logs.Normal(pattern.LOGO)
				
				defer dbutils.CloseDB()

				distriWrapper, hwInfo, err := control.GetDistri(true)
				if err != nil {
					logs.Error(fmt.Sprintf("GetDistri: %v", err))
					return nil
				}

				if err = nginx.StartNginx(
					config.GlobalConfig.Console.DistriPort,
					config.GlobalConfig.Console.WorkPort,
					config.GlobalConfig.Console.ServerPort); err != nil {
					logs.Error(fmt.Sprintf("StartNginx error: %v", err))
					return nil
				}

				machine, err := distriWrapper.GetMachine()
				if err != nil {
					logs.Error(fmt.Sprintf("GetMachine: %v", err))
					return nil
				}

				if machine.Metadata == "" {
					logs.Normal("Machine does not exist")
					_, err := distriWrapper.AddMachine(*hwInfo)
					if err != nil {
						logs.Error(fmt.Sprintf("AddMachine: %v", err))
						return nil
					}
				} else {
					logs.Normal("Machine already exists")
				}

				go server.StartServer(config.GlobalConfig.Console.ServerPort)

				control.StartHeartbeatTask(distriWrapper, hwInfo.MachineUUID)

				for {
					time.Sleep(1 * time.Minute)

					machine, err = distriWrapper.GetMachine()
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
						// TODO: Add the logic of the ForRent status.
						break ListenLoop
					case "Renting":

						logs.Normal(fmt.Sprintf("Machine is Renting, Details: %v", machine))

						orderID := machine.OrderPda
						if orderID.Equals(solana.SystemProgramID) {
							logs.Error(fmt.Sprintf("machine OrderPda error, OrderPda: %v", orderID))
							break ListenLoop
						}

						distriWrapper.ProgramDistriOrder = orderID
						newOrder, err := distriWrapper.GetOrder()
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
							// orderPlacedMetadata.OrderInfo.Message = "This is a test text.This is a test text.This is a test text.This is a test text.This is a test text."
							// if err = control.OrderFailed(distriWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
							// 	logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
							// }
							// break ListenLoop

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
								if err = control.OrderFailed(distriWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
									logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
								}
								break ListenLoop
							}

							url := orderPlacedMetadata.OrderInfo.DownloadURL
							if len(url) > 0 {
								modelDir := config.GlobalConfig.Console.WorkDirectory + "/ml-workspace"
								var modelURL []utils.DownloadURL

								// Easy debugging
								for _, u := range url {
									modelURL = append(modelURL, utils.DownloadURL{
										URL: config.GlobalConfig.Console.IpfsNodeUrl + "/ipfs" + utils.EnsureLeadingSlash(u),
										// URL:      u,
										Checksum: "",
										Name:     "CID.json",
									})
								}

								logs.Normal("Downloading CID.json ...")
								err = utils.DownloadFiles(modelDir, modelURL)
								if err != nil {
									logs.Error(fmt.Sprintf("DownloadFiles %v", err))
								}

								items, err := utils.GetCidItemsFromFile(modelDir + "/CID.json")
								if err != nil {
									logs.Error(fmt.Sprintf("GetCidItemsFromFile %v", err))
								}

								modelURL = nil
								for _, item := range items {
									modelURL = append(modelURL, utils.DownloadURL{
										URL:      config.GlobalConfig.Console.IpfsNodeUrl + "/ipfs" + utils.EnsureLeadingSlash(item.Cid),
										Checksum: "",
										Name:     item.Name,
									})
								}

								logs.Normal("Downloading the following files...")
								for _, url := range modelURL {
									logs.Normal(url.Name)
								}

								err = utils.DownloadFiles(modelDir, modelURL)
								if err != nil {
									logs.Error(fmt.Sprintf("DownloadFiles %v", err))
								}
							}
						case "deploy":
							_, err := dbutils.GenToken(newOrder.Buyer.String())
							if err != nil {
								logs.Error(fmt.Sprintf("GenToken: %v", err))
								break ListenLoop
							}

							// Easy debugging
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
								if err = control.OrderFailed(distriWrapper, orderPlacedMetadata, newOrder.Buyer); err != nil {
									logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
								}
								break ListenLoop
							}
							// containerID, err = docker.RunDeployContainer(isGPU, orderPlacedMetadata.OrderInfo.DownloadURL)
							// if err != nil {
							// 	logs.Error(fmt.Sprintln("RunDeployContainer error ", err))

							// 	if err = control.OrderFailed(distriWrapper, newOrder.Metadata, newOrder.Buyer); err != nil {
							// 		logs.Error(fmt.Sprintf("control.OrderFailed: %v", err))
							// 	}
							// 	break ListenLoop
							// }
						default:
							logs.Error(fmt.Sprintf("OrderInfo.Intent error, Intent: %v", orderPlacedMetadata.OrderInfo.Intent))
							break ListenLoop
						}

						_, err = distriWrapper.OrderStart()
						if err != nil {
							logs.Error(fmt.Sprintf("OrderStart: %v", err))
							if err := docker.StopWorkspaceContainer(containerID); err != nil {
								logs.Error(fmt.Sprintf("> StopWorkspaceContainer, containerID: %s, err: %v", containerID, err))
							}
							break ListenLoop
						}

						for {
							time.Sleep(1 * time.Minute)

							newOrder, err = distriWrapper.GetOrder()
							if err != nil {
								logs.Error(fmt.Sprintf("GetOrder Error: %v", err))
								break ListenLoop
							}

							switch newOrder.Status.String() {
							case "Preparing":
								logs.Error(fmt.Sprintf("Order error, ID: %v\norder: %v", distriWrapper.ProgramDistriOrder, newOrder))
								break ListenLoop
							case "Training":
								orderEndTime := time.Unix(newOrder.StartTime, 0).Add(time.Hour * time.Duration(newOrder.Duration))

								db := dbutils.GetDB()
								dbutils.Update(db, []byte("orderEndTime"), []byte(orderEndTime.Format(time.RFC3339)))

								timeNow := time.Now()
								if timeNow.After(orderEndTime) {

									logs.Normal(fmt.Sprintf("Order completed, Details: %v", newOrder))

									if err = control.OrderComplete(distriWrapper, newOrder.Metadata, isGPU, containerID); err != nil {
										logs.Error(fmt.Sprintf("OrderComplete: %v", err))
									}
									break ListenLoop
								}
								continue
							case "Completed":
								logs.Error(fmt.Sprintf("Order error, ID: %v\norder: %v", distriWrapper.ProgramDistriOrder, newOrder))
								break ListenLoop
							case "Failed":
								logs.Error(fmt.Sprintf("Order error, ID: %v\norder: %v", distriWrapper.ProgramDistriOrder, newOrder))
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

				distriWrapper, _, err := control.GetDistri(false)
				if err != nil {
					logs.Error(err.Error())
					return nil
				}

				hash, err := distriWrapper.RemoveMachine()
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
