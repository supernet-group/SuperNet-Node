package super

import (
	"SuperNet-Node/chain"
	"SuperNet-Node/chain/super/supernet"
	"SuperNet-Node/docker"
	"SuperNet-Node/machine_info"
	"SuperNet-Node/pattern"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type WrapperSuper struct {
	*chain.InfoChain
}

// Register the given hardware information with a distributed system or blockchain
func (chain WrapperSuper) AddMachine(hardwareInfo machine_info.MachineInfo) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_REGISTER))

	latest, err := chain.Conn.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("GetLatestBlockhash error: %s", err)
	}

	uuid, err := utils.ParseMachineUUID(string(hardwareInfo.MachineUUID))
	if err != nil {
		return "", fmt.Errorf("> ParseMachineUUID: %v", err.Error())
	}

	jsonData, err := json.Marshal(hardwareInfo)
	if err != nil {
		return "", fmt.Errorf("> json.Marshal: %v", err.Error())
	}

	seedStatisticsOwner := utils.GenStatisticsOwner(chain.Wallet.Wallet.PublicKey())
	statisticsOwner, _, err := solana.FindProgramAddress(
		seedStatisticsOwner,
		chain.ProgramSuperID,
	)
	if err != nil {
		return "", fmt.Errorf("> FindProgramAddress: %v", err)
	}

	supernet.SetProgramID(chain.ProgramSuperID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			supernet.NewAddMachineInstruction(
				uuid,
				string(jsonData),
				chain.ProgramSuperMachine,
				chain.Wallet.Wallet.PublicKey(),
				statisticsOwner,
				solana.SystemProgramID,
			).Build(),
		},
		latest.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("> NewAddMachineInstruction: %v", err.Error())
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("> tx.Sign: %v", err.Error())
	}

	logs.Normal("=============== AddMachine Transaction")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_REGISTER, sig))

	return sig, nil
}

func (chain WrapperSuper) RemoveMachine() (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %s", pattern.TX_HASHRATE_MARKET_REMOVE_MACHINE))

	latest, err := chain.Conn.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	supernet.SetProgramID(chain.ProgramSuperID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			supernet.NewRemoveMachineInstruction(
				chain.ProgramSuperMachine,
				chain.Wallet.Wallet.PublicKey(),
			).Build(),
		},
		latest.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== RemoveMachine Transaction")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_REMOVE_MACHINE, sig))

	return sig, nil
}

func (chain WrapperSuper) OrderStart() (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_START))

	latest, err := chain.Conn.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("> GetLatestBlockhash: %v", err)
	}

	supernet.SetProgramID(chain.ProgramSuperID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			supernet.NewStartOrderInstruction(
				chain.ProgramSuperOrder,
				chain.Wallet.Wallet.PublicKey(),
			).Build(),
		},
		latest.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("> solana.NewTransaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("> tx.Sign: %v", err)
	}

	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_START, sig))

	return sig, nil
}

func (chain WrapperSuper) OrderCompleted(order supernet.Order, isGPU bool) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_COMPLETED))

	score, err := docker.RunScoreContainer(isGPU)
	if err != nil {
		return "", err
	}
	scoreUint8 := uint8(score)

	latest, err := chain.Conn.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	seller := chain.Wallet.Wallet.PublicKey()
	mint := solana.MustPublicKeyFromBase58(pattern.SNT_TOKEN_ID)
	sellerAta, _, err := solana.FindAssociatedTokenAddress(seller, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	seedVault := utils.GenVault()
	vault, _, err := solana.FindProgramAddress(
		seedVault,
		chain.ProgramSuperID,
	)
	if err != nil {
		return "", fmt.Errorf("error finding program address: %v", err)
	}

	var model1OwnerAta, model2OwnerAta, model3OwnerAta, model4OwnerAta, model5OwnerAta,
		statisticsModel1Owner, statisticsModel2Owner, statisticsModel3Owner, statisticsModel4Owner, statisticsModel5Owner solana.PublicKey = chain.ProgramSuperID,
		chain.ProgramSuperID, chain.ProgramSuperID, chain.ProgramSuperID, chain.ProgramSuperID, chain.ProgramSuperID,
		chain.ProgramSuperID, chain.ProgramSuperID, chain.ProgramSuperID, chain.ProgramSuperID

	if !order.Model1Owner.IsZero() {
		model1OwnerAta, _, err = solana.FindAssociatedTokenAddress(order.Model1Owner, mint)
		if err != nil {
			return "", fmt.Errorf("error finding model1Owner associated token address: %v", err)
		}
		statisticsModel1Owner, _, err = solana.FindProgramAddress(
			utils.GenStatisticsOwner(model1OwnerAta),
			chain.ProgramSuperID,
		)
		if err != nil {
			return "", fmt.Errorf("> FindProgramAddress: %v", err)
		}
	}
	if !order.Model2Owner.IsZero() {
		model2OwnerAta, _, err = solana.FindAssociatedTokenAddress(order.Model2Owner, mint)
		if err != nil {
			return "", fmt.Errorf("error finding model2Owner associated token address: %v", err)
		}
		statisticsModel2Owner, _, err = solana.FindProgramAddress(
			utils.GenStatisticsOwner(model2OwnerAta),
			chain.ProgramSuperID,
		)
		if err != nil {
			return "", fmt.Errorf("> FindProgramAddress: %v", err)
		}
	}
	if !order.Model3Owner.IsZero() {
		model3OwnerAta, _, err = solana.FindAssociatedTokenAddress(order.Model3Owner, mint)
		if err != nil {
			return "", fmt.Errorf("error finding model3Owner associated token address: %v", err)
		}
		statisticsModel3Owner, _, err = solana.FindProgramAddress(
			utils.GenStatisticsOwner(model3OwnerAta),
			chain.ProgramSuperID,
		)
		if err != nil {
			return "", fmt.Errorf("> FindProgramAddress: %v", err)
		}
	}
	if !order.Model4Owner.IsZero() {
		model4OwnerAta, _, err = solana.FindAssociatedTokenAddress(order.Model4Owner, mint)
		if err != nil {
			return "", fmt.Errorf("error finding model1Owner associated token address: %v", err)
		}
		statisticsModel4Owner, _, err = solana.FindProgramAddress(
			utils.GenStatisticsOwner(model4OwnerAta),
			chain.ProgramSuperID,
		)
		if err != nil {
			return "", fmt.Errorf("> FindProgramAddress: %v", err)
		}
	}
	if !order.Model5Owner.IsZero() {
		model5OwnerAta, _, err = solana.FindAssociatedTokenAddress(order.Model5Owner, mint)
		if err != nil {
			return "", fmt.Errorf("error finding model1Owner associated token address: %v", err)
		}
		statisticsModel5Owner, _, err = solana.FindProgramAddress(
			utils.GenStatisticsOwner(model5OwnerAta),
			chain.ProgramSuperID,
		)
		if err != nil {
			return "", fmt.Errorf("> FindProgramAddress: %v", err)
		}
	}

	seedStatisticsOwner := utils.GenStatisticsOwner(chain.Wallet.Wallet.PublicKey())
	statisticsOwner, _, err := solana.FindProgramAddress(
		seedStatisticsOwner,
		chain.ProgramSuperID,
	)
	if err != nil {
		return "", fmt.Errorf("> FindProgramAddress: %v", err)
	}

	supernet.SetProgramID(chain.ProgramSuperID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			supernet.NewOrderCompletedInstruction(
				order.Metadata,
				scoreUint8,
				chain.ProgramSuperMachine,
				chain.ProgramSuperOrder,
				seller,
				sellerAta,
				model1OwnerAta,
				model2OwnerAta,
				model3OwnerAta,
				model4OwnerAta,
				model5OwnerAta,
				statisticsOwner,
				statisticsModel1Owner,
				statisticsModel2Owner,
				statisticsModel3Owner,
				statisticsModel4Owner,
				statisticsModel5Owner,
				vault,
				mint,
				solana.TokenProgramID,
				solana.SPLAssociatedTokenAccountProgramID,
				solana.SystemProgramID,
			).Build(),
		},
		latest.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== OrderCompleted Transaction ==================")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_COMPLETED, sig))

	return sig, nil
}

func (chain WrapperSuper) OrderFailed(buyer solana.PublicKey, orderPlacedMetadata pattern.OrderPlacedMetadata) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_FAILED))

	latest, err := chain.Conn.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	jsonData, err := json.Marshal(orderPlacedMetadata)
	if err != nil {
		return "", fmt.Errorf("> json.Marshal: %v", err.Error())
	}

	seller := chain.Wallet.Wallet.PublicKey()
	ecpc := solana.MustPublicKeyFromBase58(pattern.SNT_TOKEN_ID)
	buyerAta, _, err := solana.FindAssociatedTokenAddress(buyer, ecpc)
	if err != nil {
		return "", fmt.Errorf("> FindAssociatedTokenAddress: %v", err.Error())
	}

	seedVault := utils.GenVault()
	vault, _, err := solana.FindProgramAddress(
		seedVault,
		chain.ProgramSuperID,
	)
	if err != nil {
		return "", fmt.Errorf("> FindProgramAddress: %v", err.Error())
	}

	supernet.SetProgramID(chain.ProgramSuperID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			supernet.NewOrderFailedInstruction(
				string(jsonData),
				chain.ProgramSuperMachine,
				chain.ProgramSuperOrder,
				seller,
				buyerAta,
				vault,
				ecpc,
				solana.TokenProgramID,
				solana.SPLAssociatedTokenAccountProgramID,
			).Build(),
		},
		latest.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("> NewOrderFailedInstruction: %v", err.Error())
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("> tx.Sign: %v", err.Error())
	}

	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_FAILED, sig))

	return sig, nil
}

func (chain WrapperSuper) OrderRefund(buyer solana.PublicKey) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_REFUND))

	latest, err := chain.Conn.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	seller := chain.Wallet.Wallet.PublicKey()
	mint := solana.MustPublicKeyFromBase58(pattern.SNT_TOKEN_ID)
	sellerAta, _, err := solana.FindAssociatedTokenAddress(seller, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}
	buyerAta, _, err := solana.FindAssociatedTokenAddress(buyer, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	seedVault := utils.GenVault()
	vault, _, err := solana.FindProgramAddress(
		seedVault,
		chain.ProgramSuperID,
	)
	if err != nil {
		return "", fmt.Errorf("error finding program address: %v", err)
	}

	seedStatisticsOwner := utils.GenStatisticsOwner(chain.Wallet.Wallet.PublicKey())
	statisticsOwner, _, err := solana.FindProgramAddress(
		seedStatisticsOwner,
		chain.ProgramSuperID,
	)
	if err != nil {
		return "", fmt.Errorf("> FindProgramAddress: %v", err)
	}

	// zeroPublicKey := solana.MustPublicKeyFromBase58("11111111111111111111111111111111")
	zeroPublicKey := chain.ProgramSuperID

	supernet.SetProgramID(chain.ProgramSuperID)

	refundOrder := supernet.NewRefundOrderInstructionBuilder()
	refundOrder.AccountMetaSlice[0] = solana.Meta(chain.ProgramSuperMachine).WRITE()
	refundOrder.AccountMetaSlice[1] = solana.Meta(chain.ProgramSuperOrder).WRITE()
	refundOrder.AccountMetaSlice[2] = solana.Meta(buyer).WRITE()
	refundOrder.AccountMetaSlice[3] = solana.Meta(buyerAta).WRITE()
	refundOrder.AccountMetaSlice[4] = solana.Meta(sellerAta).WRITE()
	refundOrder.AccountMetaSlice[5] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[6] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[7] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[8] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[9] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[10] = solana.Meta(statisticsOwner).WRITE()
	refundOrder.AccountMetaSlice[11] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[12] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[13] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[14] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[15] = solana.Meta(zeroPublicKey).WRITE()
	refundOrder.AccountMetaSlice[16] = solana.Meta(vault).WRITE()
	refundOrder.AccountMetaSlice[17] = solana.Meta(mint).WRITE()
	refundOrder.AccountMetaSlice[18] = solana.Meta(solana.TokenProgramID).WRITE()
	refundOrder.AccountMetaSlice[19] = solana.Meta(solana.SPLAssociatedTokenAccountProgramID).WRITE()
	refundOrder.AccountMetaSlice[20] = solana.Meta(solana.SystemProgramID).WRITE()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			refundOrder.Build(),
		},
		latest.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== OrderRefund Transaction ==================")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_REFUND, sig))

	return sig, nil
}

func (chain WrapperSuper) GetMachine() (supernet.Machine, error) {

	var data supernet.Machine

	resp, err := chain.Conn.RpcClient.GetAccountInfo(
		context.TODO(),
		chain.ProgramSuperMachine,
	)
	if err != nil {
		return data, nil
	}

	borshDec := bin.NewBorshDecoder(resp.GetBinary())

	err = data.UnmarshalWithDecoder(borshDec)
	if err != nil {
		return data, fmt.Errorf("> UnmarshalWithDecoder: %v", err)
	}

	return data, nil
}

func (chain WrapperSuper) GetOrder() (supernet.Order, error) {

	var data supernet.Order

	resp, err := chain.Conn.RpcClient.GetAccountInfo(
		context.TODO(),
		chain.ProgramSuperOrder,
	)
	if err != nil {
		return data, nil
	}

	borshDec := bin.NewBorshDecoder(resp.GetBinary())

	err = data.UnmarshalWithDecoder(borshDec)
	if err != nil {
		return data, fmt.Errorf("error unmarshaling data: %v", err)
	}

	return data, nil
}

func (chain WrapperSuper) SubmitTask(
	taskUuid pattern.TaskUUID,
	machineUUID pattern.MachineUUID,
	period uint32,
	taskMetadata pattern.TaskMetadata) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_SUBMIT_TASK))

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error getting recent blockhash: %v", err)
	}

	jsonData, err := json.Marshal(taskMetadata)
	if err != nil {
		return "", fmt.Errorf("error marshaling the struct to JSON: %v", err)
	}

	programID := solana.MustPublicKeyFromBase58(pattern.PROGRAM_SUPER_ID)
	seedTask := utils.GenTask(chain.Wallet.Wallet.PublicKey(), taskUuid)
	task, _, _ := solana.FindProgramAddress(
		seedTask,
		programID,
	)
	seedReward := utils.GenReward()
	reward, _, _ := solana.FindProgramAddress(
		seedReward,
		programID,
	)
	seedRewardMachine := utils.GenRewardMachine(chain.Wallet.Wallet.PublicKey(), machineUUID)
	rewardMachine, _, _ := solana.FindProgramAddress(
		seedRewardMachine,
		programID,
	)

	supernet.SetProgramID(chain.ProgramSuperID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			supernet.NewSubmitTaskInstruction(
				taskUuid,
				utils.CurrentPeriod(),
				string(jsonData),
				chain.ProgramSuperMachine,
				task,
				reward,
				rewardMachine,
				chain.Wallet.Wallet.PublicKey(),
				solana.SystemProgramID,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_SUBMIT_TASK, sig))

	return sig, nil
}

func NewSuperWrapper(info *chain.InfoChain) *WrapperSuper {
	return &WrapperSuper{info}
}
