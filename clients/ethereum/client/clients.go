package client

import (
	"context"
	"crypto-trade-client/clients/ethereum"
	"crypto-trade-client/common/rpc"
	"crypto-trade-client/common/stringutil"
)

type EthClient struct {
	ethRpc *ethereum.EthRpc
}

// NewEthClient 创建并初始化一个新的 EthClient
func NewEthClient(endpoint string, chainName string) (*EthClient, error) {
	var eRPC ethereum.EthRpc
	err := rpc.NewClient(context.Background(), endpoint, chainName, &eRPC, map[string]string{})
	if err != nil {
		return nil, err
	}

	return &EthClient{ethRpc: &eRPC}, nil
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
