package core

import (
	ethclient "crypto-trade-client/clients/ethereum/client"
)

type CoreClient struct {
	*ethclient.EthClient
}

// NewCoreClient creates a new Core client with the given endpoint and private key
func NewCoreClient(endpoint string, privateKey string) (*CoreClient, error) {
	client, err := ethclient.NewEthClient(endpoint, "core", privateKey)
	if err != nil {
		return nil, err
	}

	return &CoreClient{client}, nil
}
