package wallet

import (
	"SuperNet-Node/config"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

type Wallet struct {
	Wallet *solana.Wallet
}

func InitWallet(cfg *config.SolanaConfig) (*Wallet, error) {
	payerPrivateKey := solana.MustPrivateKeyFromBase58(cfg.Key)
	payerAccount, err := solana.WalletFromPrivateKeyBase58(payerPrivateKey.String())

	if err != nil {
		return nil, fmt.Errorf("> WalletFromPrivateKeyBase58: %v", err)
	}

	wallet := &Wallet{
		Wallet: payerAccount,
	}

	return wallet, nil
}
