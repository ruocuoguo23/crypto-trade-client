package client

import (
	"context"
	"crypto-trade-client/clients/ethereum"
	"crypto-trade-client/common/rpc"
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
