package client

import (
	"context"
	"crypto-trade-client/clients/ethereum"
	"crypto-trade-client/common/rpc"
	"crypto-trade-client/common/stringutil"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type EvmSigner struct {
	PrivateKey    *ecdsa.PrivateKey
	PublicAddress common.Address
}

type EthClient struct {
	ethRpc *ethereum.EthRpc

	ethClient *ethclient.Client

	//privateKey *ecdsa.PrivateKey
	signer *EvmSigner
}

// NewEthClient creates a new Ethereum client with the given endpoint, chain name, and private key
func NewEthClient(endpoint string, chainName string, privateKeyHex string) (*EthClient, error) {
	var eRPC ethereum.EthRpc
	err := rpc.NewClient(context.Background(), endpoint, chainName, &eRPC, map[string]string{})
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	if stringutil.IsBlank(privateKeyHex) {
		// no private key provided, so we can't sign transactions
		client := &EthClient{
			ethRpc:    &eRPC,
			ethClient: ethClient,
			signer:    nil,
		}

		return client, nil
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	signer := &EvmSigner{
		PrivateKey:    privateKey,
		PublicAddress: crypto.PubkeyToAddress(*publicKeyECDSA),
	}

	client := &EthClient{
		ethRpc:    &eRPC,
		ethClient: ethClient,
		signer:    signer,
	}

	return client, nil
}

func (ec *EthClient) GetBlock(hash string, height int64) (*ethereum.Block, error) {
	bb, err := ec.bestBlockHeader()
	if err != nil {
		return nil, err
	}
	if height > bb.Height {
		return nil, ethereum.ErrBlockNotFound
	}

	var block ethereum.Block
	if hash != "" {
		block, err = ec.ethRpc.GetBlockByHash(hash, true)
		if err != nil {
			return nil, err
		}
		if stringutil.IsBlank(block.Hash) {
			return nil, ethereum.ErrBlockNotFound
		}
	} else {
		if height > bb.Height {
			return nil, ethereum.ErrBlockNotFound
		}
		block, err = ec.ethRpc.GetBlockByNumber(ethereum.EthBlockNumArg(height), true)
		if err != nil {
			return nil, err
		}
	}

	return &block, nil
}

func (ec *EthClient) GetLatestBlockHeight() (int64, error) {
	bh, err := ec.bestBlockHeader()
	if err != nil {
		return 0, err
	}
	return bh.Height, nil
}

func (ec *EthClient) bestBlockHeader() (ethereum.BlockHeader, error) {
	block, err := ec.ethRpc.GetBlockByNumber(ethereum.Latest, false)
	if err != nil {
		return ethereum.BlockHeader{}, err
	}
	bh := ethereum.BlockHeader{
		Hash:   block.Hash,
		Prev:   block.ParentHash,
		Height: int64(block.Number),
		Time:   int64(block.Time),
	}

	return bh, nil
}

func (ec *EthClient) GetChainID(ctx context.Context) (*big.Int, error) {
	result, err := ec.ethClient.ChainID(ctx)
	return result, err
}

func (ec *EthClient) PendingNonceAt(ctx context.Context, address common.Address) (uint64, error) {
	return ec.ethClient.PendingNonceAt(ctx, address)
}

func (ec *EthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return ec.ethClient.SuggestGasPrice(ctx)
}

func (ec *EthClient) _getSinnerPrivateKey(signer common.Address) (*ecdsa.PrivateKey, error) {
	if ec.signer.PublicAddress != signer {
		return nil, errors.New("signer address does not match")
	}
	return ec.signer.PrivateKey, nil
}

// Transfer sends amount of ether or token to the given address
func (ec *EthClient) Transfer(signer common.Address, to common.Address, value *big.Int) (common.Hash, error) {
	ctx := context.Background()

	msgSignerPk, err := ec._getSinnerPrivateKey(signer)
	if err != nil {
		return common.Hash{}, err
	}

	// nonce
	nonce, err := ec.PendingNonceAt(ctx, signer)
	if err != nil {
		return common.Hash{}, err
	}

	// gas limit
	gasLimit := uint64(21000)

	// gas price
	gasPrice, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	chainID, err := ec.GetChainID(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	signedTx, err := types.SignNewTx(msgSignerPk, types.NewEIP155Signer(chainID), &types.LegacyTx{
		To:       &to,
		Nonce:    nonce,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
	})
	if err != nil {
		return common.Hash{}, err
	}

	err = ec.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}
