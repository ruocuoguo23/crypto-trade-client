package client

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

type EthSigner struct {
	privateKey *ecdsa.PrivateKey
}

func NewEthSigner(hexKey string) (*EthSigner, error) {
	privateKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, err
	}
	return &EthSigner{privateKey: privateKey}, nil
}

func (s *EthSigner) SignTransaction(tx interface{}) ([]byte, error) {
	// Implement Ethereum transaction signing logic here
	// For demonstration purposes, we assume tx is a byte slice
	data, ok := tx.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type")
	}
	signature, err := crypto.Sign(data, s.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}
