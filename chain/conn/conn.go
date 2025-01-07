package conn

import (
	"SuperNet-Node/config"
	logs "SuperNet-Node/utils/log_utils"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// Connection to Solana Blockchain Nodes
type Conn struct {
	RpcClient *rpc.Client
}

// Receive a SolanaConfig configuration object and return an initialized Conn object.
func NewConn(cfg *config.SolanaConfig) (*Conn, error) {

	rpcClient := rpc.New(cfg.RPC)

	conn := &Conn{
		RpcClient: rpcClient,
	}

	return conn, nil
}

func (conn *Conn) SendAndConfirmTransaction(tx *solana.Transaction) (string, error) {
	sig, err := conn.RpcClient.SendTransactionWithOpts(context.TODO(), tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized})
	if err != nil {
		return "", fmt.Errorf("> SendTransactionWithOpts, txSig: %s, err: %v", sig.String(), err)
	}

	_, err = conn.WaitForConfirmation(sig.String())
	if err != nil {
		return "", fmt.Errorf("> WaitForConfirmation: %v", err)
	}

	return sig.String(), nil
}

const TryLimit = 10

func (conn *Conn) WaitForConfirmation(tx string) (*solana.Transaction, error) {
	var transaction *solana.Transaction
	txSig := solana.MustSignatureFromBase58(tx)

	var tryNum = 0
	for {
		tryNum++
		if tryNum > TryLimit {
			return nil, fmt.Errorf("> TryLimit reached, tx: %s", tx)
		}
		time.Sleep(3 * time.Second)

		out, err := conn.RpcClient.GetTransaction(
			context.TODO(),
			txSig,
			nil,
		)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			return nil, fmt.Errorf("> RpcClient.GetTransaction, tx: %s, err: %v", tx, err)
		}

		if out == nil {
			logs.Warning("out is nil")
			continue
		}

		if out.Meta.Err != nil {
			return nil, fmt.Errorf("> out.Meta.Err: %v", out.Meta.Err)
		}

		transaction, err = out.Transaction.GetTransaction()
		if err != nil {
			return nil, fmt.Errorf("> Transaction.GetTransaction, tx: %s, err: %v", tx, err)
		}

		if transaction != nil {
			break
		}
	}
	return transaction, nil
}
