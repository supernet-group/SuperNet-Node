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

// SendAndConfirmTransaction sends a transaction via the RPC client and waits for its confirmation.
func (conn *Conn) SendAndConfirmTransaction(tx *solana.Transaction) (string, error) {
	// Sends the transaction and acquires its signature
	sig, err := conn.RpcClient.SendTransactionWithOpts(context.TODO(), tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized})
	if err != nil {
		// Returns an error if the transaction fails to send
		return "", fmt.Errorf("> SendTransactionWithOpts, txSig: %s, err: %v", sig.String(), err)
	}

	// Waits for the transaction to be confirmed by the network
	_, err = conn.WaitForConfirmation(sig.String())
	if err != nil {
		// Returns an error if confirmation is not successful
		return "", fmt.Errorf("> WaitForConfirmation: %v", err)
	}

	// Returns the transaction signature upon successful confirmation
	return sig.String(), nil
}

const TryLimit = 10

// WaitForConfirmation waits for confirmation of a Solana transaction.
func (conn *Conn) WaitForConfirmation(tx string) (*solana.Transaction, error) {
	var transaction *solana.Transaction
	txSig := solana.MustSignatureFromBase58(tx)

	var tryNum = 0
	for {
		tryNum++
		// If the number of attempts exceeds the limit, return an error.
		if tryNum > TryLimit {
			return nil, fmt.Errorf("> TryLimit reached, tx: %s", tx)
		}
		// Sleep for 3 seconds between each attempt.
		time.Sleep(3 * time.Second)

		// Call the RPC client to get transaction details.
		out, err := conn.RpcClient.GetTransaction(
			context.TODO(),
			txSig,
			nil,
		)
		// If the transaction is not found, continue trying.
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			// Return other errors.
			return nil, fmt.Errorf("> RpcClient.GetTransaction, tx: %s, err: %v", tx, err)
		}

		// Handle the case where 'out' is unexpectedly nil.
		if out == nil {
			logs.Warning("out is nil")
			continue
		}

		// If there's an error in the transaction metadata, return it.
		if out.Meta.Err != nil {
			return nil, fmt.Errorf("> out.Meta.Err: %v", out.Meta.Err)
		}

		// Retrieve the transaction details; if successful, break the loop.
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
