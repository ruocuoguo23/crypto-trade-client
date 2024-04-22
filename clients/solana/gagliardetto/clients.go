package gagliardetto

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolClient struct {
	ctx     context.Context
	account *solana.Wallet
	client  *rpc.Client
	keyPath string
}

func NewSolClient(ctx context.Context, keyPath string) (*SolClient, error) {
	account, err := loadWallet(keyPath)
	if err != nil {
		return nil, err
	}

	// Create an RPC client instance
	c := rpc.New(rpc.DevNet_RPC)

	return &SolClient{
		ctx:     ctx,
		account: account,
		client:  c,
		keyPath: keyPath,
	}, nil
}

func (c *SolClient) GetWallet() *solana.Wallet {
	return c.account
}

func (c *SolClient) SendTransaction(instruction solana.Instruction) (string, error) {
	recentBlockHash, err := c.client.GetRecentBlockhash(c.ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recentBlockHash.Value.Blockhash,
	)
	if err != nil {
		return "", err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if c.account.PublicKey().Equals(key) {
				return &c.account.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", err
	}

	sig, err := c.client.SendTransaction(c.ctx, tx)
	if err != nil {
		return "", err
	}

	return sig.String(), nil
}

func (c *SolClient) GetTransaction(txHash solana.Signature) (*solana.Transaction, error) {
	out, err := c.client.GetTransaction(c.ctx, txHash, nil)
	if err != nil {
		return nil, err
	}

	return out.Transaction.GetTransaction()
}

func (c *SolClient) GetBalance() (uint64, error) {
	balance, err := c.client.GetBalance(c.ctx, c.account.PublicKey(), rpc.CommitmentConfirmed)
	if err != nil {
		return 0, err
	}
	return balance.Value, nil
}

func loadWallet(keyPath string) (*solana.Wallet, error) {
	privateKey, err := solana.PrivateKeyFromSolanaKeygenFile(keyPath)
	if err != nil {
		return nil, err
	}

	return &solana.Wallet{PrivateKey: privateKey}, nil
}
