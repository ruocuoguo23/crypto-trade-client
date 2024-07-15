package client

import (
	ethclient "crypto-trade-client/clients/ethereum/client"
)

type OptimismClient struct {
	*ethclient.EthClient
}

// NewOpClient creates a new Optimism client with the given endpoint and private key
func NewOpClient(endpoint string, privateKey string) (*OptimismClient, error) {
	client, err := ethclient.NewEthClient(endpoint, "optimism", privateKey)
	if err != nil {
		return nil, err
	}

	return &OptimismClient{client}, nil
}
