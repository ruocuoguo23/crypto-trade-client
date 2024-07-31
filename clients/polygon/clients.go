package polygon

import (
	ethclient "crypto-trade-client/clients/ethereum/client"
)

type PolyClient struct {
	*ethclient.EthClient
}

// NewPolyClient creates a new Polygon client with the given endpoint and private key
func NewPolyClient(endpoint string, privateKey string) (*PolyClient, error) {
	client, err := ethclient.NewEthClient(endpoint, "core", privateKey)
	if err != nil {
		return nil, err
	}

	return &PolyClient{client}, nil
}
