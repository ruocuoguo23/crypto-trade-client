package main

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

func loadAccount() (*types.Account, error) {
	// Load a keypair from local file
	filePath := "/Users/jeff.wu/.config/solana/id.json"
	fileContent, err := os.ReadFile(filePath)
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

func main() {
	// Create an RPC client instance
	// temp using local network endpoint
	c := client.NewClient(rpc.LocalnetRPCEndpoint)

	// Load the account
	account, err := loadAccount()
	if err != nil {
		fmt.Println("load account error", err)
		return
	}

	// Program ID
	programID := common.PublicKeyFromString("FyCJ7kDf2RbfoXpuCKT1KhKQxhgbgb9Wj9esDYrm1K6h")

	// Get the most latest blockhash
	res, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		fmt.Println("get latest blockhash error", err)
		return
	}

	// Build the transaction and instruction
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{*account},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        account.PublicKey,
			RecentBlockhash: res.Blockhash,
			Instructions: []types.Instruction{
				{
					ProgramID: programID,
					Accounts:  []types.AccountMeta{},
					Data:      []byte{},
				},
			},
		}),
	})

	if err != nil {
		fmt.Println("build transaction error", err)
		return
	}

	// Sign and send the transaction
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("send transaction error", err)
		return
	}

	fmt.Printf("txhash: %s\n", txhash)
}
