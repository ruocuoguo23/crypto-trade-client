package client

import (
	ethclient "crypto-trade-client/clients/ethereum/client"
)

type OptimismClient struct {
	*ethclient.EthClient
}

// NewOpClient 创建并初始化一个新的 EthClient
func NewOpClient(endpoint string) (*OptimismClient, error) {
	client, err := ethclient.NewEthClient(endpoint, "optimism")
	if err != nil {
		return nil, err
	}

	return &OptimismClient{client}, nil
}
