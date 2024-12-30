package chain

import (
	"SuperNet-Node/chain/conn"
	"SuperNet-Node/chain/wallet"
	"SuperNet-Node/config"
	"SuperNet-Node/machine_info/machine_uuid"
	"SuperNet-Node/pattern"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// InfoChain encapsulates basic information about the chain
type InfoChain struct {
	Conn                *conn.Conn
	Wallet              *wallet.Wallet
	ProgramSuperID      solana.PublicKey
	ProgramSuperMachine solana.PublicKey
	ProgramSuperOrder   solana.PublicKey
}

// GetChainInfo returns *Infochain and error when the connection fails
func GetChainInfo(cfg *config.SolanaConfig, machineUUID machine_uuid.MachineUUID) (*InfoChain, error) {
	newConn, err := conn.NewConn(cfg)
	if err != nil {
		return nil, fmt.Errorf("> conn.NewConn: %v", err)
	}

	wallet, err := wallet.InitWallet(cfg)
	if err != nil {
		return nil, fmt.Errorf("> wallet.InitWallet: %v", err)
	}

	programID := solana.MustPublicKeyFromBase58(pattern.PROGRAM_SUPER_ID)

	seedMachine := utils.GenMachine(wallet.Wallet.PublicKey(), machineUUID)

	machineAccount, _, err := solana.FindProgramAddress(
		seedMachine,
		programID,
	)
	if err != nil {
		return nil, fmt.Errorf("> FindProgramAddress: %v", err)
	}
	logs.Normal(fmt.Sprintf("machineAccount : %v", machineAccount.String()))

	chainInfo := &InfoChain{
		Conn:                newConn,
		Wallet:              wallet,
		ProgramSuperID:      programID,
		ProgramSuperMachine: machineAccount,
	}

	return chainInfo, nil
}
