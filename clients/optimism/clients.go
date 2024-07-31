package optimism

import (
	ethclient "crypto-trade-client/clients/ethereum/client"
)

type OpClient struct {
	*ethclient.EthClient
}

// NewOpClient creates a new Optimism client with the given endpoint and private key
func NewOpClient(endpoint string, privateKey string) (*OpClient, error) {
	client, err := ethclient.NewEthClient(endpoint, "optimism", privateKey)
	if err != nil {
		return nil, err
	}

	return &OpClient{client}, nil
}
