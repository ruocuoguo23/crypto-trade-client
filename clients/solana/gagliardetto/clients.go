package gagliardetto

import (
	"context"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolClient struct {
	ctx     context.Context
	account *solana.Wallet
	client  *rpc.Client
	keyPath string
}

func NewSolClient(ctx context.Context, keyPath string) (*SolClient, error) {
	return NewSolClientWithEndpoint(ctx, rpc.DevNet_RPC, keyPath)
}

func NewSolClientWithEndpoint(ctx context.Context, endpoint string, keyPath string) (*SolClient, error) {
	if endpoint == "" {
		return nil, errors.New("solana RPC endpoint is required")
	}
	account, err := loadWallet(keyPath)
	if err != nil {
		return nil, fmt.Errorf("load solana wallet: %w", err)
	}

	c := rpc.New(endpoint)

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

func (c *SolClient) Address() solana.PublicKey {
	return c.account.PublicKey()
}

func AddressFromKeyFile(keyPath string) (solana.PublicKey, error) {
	wallet, err := loadWallet(keyPath)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("load solana wallet: %w", err)
	}
	return wallet.PublicKey(), nil
}

func (c *SolClient) SendTransaction(instruction solana.Instruction) (string, error) {
	recentBlockHash, err := c.client.GetLatestBlockhash(c.ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("get latest solana blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recentBlockHash.Value.Blockhash,
		solana.TransactionPayer(c.account.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("build solana transaction: %w", err)
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
		return "", fmt.Errorf("sign solana transaction: %w", err)
	}

	sig, err := c.client.SendTransaction(c.ctx, tx)
	if err != nil {
		return "", fmt.Errorf("broadcast solana transaction: %w", err)
	}

	return sig.String(), nil
}

func (c *SolClient) Transfer(toAddress string, lamports uint64) (string, error) {
	if lamports == 0 {
		return "", errors.New("solana amount must be greater than zero")
	}
	recipient, err := solana.PublicKeyFromBase58(toAddress)
	if err != nil {
		return "", fmt.Errorf("invalid solana recipient: %w", err)
	}
	instruction := system.NewTransferInstruction(
		lamports,
		c.account.PublicKey(),
		recipient,
	).Build()
	return c.SendTransaction(instruction)
}

func (c *SolClient) GetTransaction(txHash string) (*solana.Transaction, error) {
	txSig := solana.MustSignatureFromBase58(txHash)
	out, err := c.client.GetTransaction(c.ctx, txSig, nil)
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
