package client

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

// EthClient 封装了与 Ethereum 节点的连接和操作
type EthClient struct {
	client *ethclient.Client
}

// NewEthClient 创建并初始化一个新的 EthClient
func NewEthClient(endpoint string) (*EthClient, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return &EthClient{client: client}, nil
}

// GetBlock 获取最新的区块并打印其内容
func (ec *EthClient) GetBlock(blockNumber int64) error {
	block, err := ec.client.BlockByNumber(context.Background(), new(big.Int).SetInt64(blockNumber))
	if err != nil {
		return err
	}

	fmt.Printf("Block number: %d\n", block.Number().Uint64())
	fmt.Printf("Block timestamp: %d\n", block.Time())
	fmt.Printf("Block transactions: %d\n", len(block.Transactions()))

	for _, tx := range block.Transactions() {
		fmt.Printf("Tx Hash: %s\n", tx.Hash().Hex())
		fmt.Printf("Tx Value: %s\n", tx.Value().String())
		fmt.Printf("Tx Gas: %d\n", tx.Gas())
		fmt.Printf("Tx Gas Price: %s\n", tx.GasPrice().String())
	}

	return nil
}
