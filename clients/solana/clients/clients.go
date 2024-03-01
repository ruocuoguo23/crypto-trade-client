package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
	"os"
)

type SolClient struct {
	ctx     context.Context
	account *types.Account
	client  *client.Client
	keyPath string
}

func NewSolClient(ctx context.Context, keyPath string) (*SolClient, error) {
	account, err := loadAccount(keyPath)
	if err != nil {
		return nil, err
	}

	// Create an RPC client instance
	// temp using local network endpoint
	c := client.NewClient(rpc.LocalnetRPCEndpoint)

	return &SolClient{
		ctx:     ctx,
		account: account,
		client:  c,
		keyPath: keyPath,
	}, nil
}

func (c *SolClient) GetAccount() *types.Account {
	return c.account
}

func (c *SolClient) SendTransaction(programID string) (string, error) {
	// Program ID
	programPK := common.PublicKeyFromString(programID)

	// Get the most latest blockhash
	res, err := c.client.GetLatestBlockhash(context.Background())
	if err != nil {
		fmt.Println("get latest block hash error", err)
		return "", err
	}

	// Build the transaction and instruction
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{*c.account},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        c.account.PublicKey,
			RecentBlockhash: res.Blockhash,
			Instructions: []types.Instruction{
				{
					ProgramID: programPK,
					Accounts:  []types.AccountMeta{},
					Data:      []byte{},
				},
			},
		}),
	})

	if err != nil {
		fmt.Println("build transaction error", err)
		return "", err
	}

	// Sign and send the transaction
	txHash, err := c.client.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("send transaction error", err)
		return txHash, err
	}

	fmt.Printf("txhash: %s\n", txHash)
	return txHash, nil
}

func (c *SolClient) GetTransaction(txHash string) (*client.Transaction, error) {
	response, err := c.client.GetTransactionWithConfig(c.ctx, txHash, client.GetTransactionConfig{
		Commitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *SolClient) GetBalance() (uint64, error) {
	// 1 SOL = 1,000,000,000 lamports
	// Get balance returns the balance in lamports
	return c.client.GetBalance(c.ctx, c.account.PublicKey.ToBase58())
}

func loadAccount(keyPath string) (*types.Account, error) {
	// Load a keypair from local file
	fileContent, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	var privateKey []byte
	err = json.Unmarshal(fileContent, &privateKey)
	if err != nil {
		return nil, err
	}

	account, err := types.AccountFromBytes(privateKey)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
