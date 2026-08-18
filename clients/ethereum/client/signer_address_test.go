package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestSignerAddress(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	want := crypto.PubkeyToAddress(privateKey.PublicKey)
	client := &EthClient{signer: &EvmSigner{PrivateKey: privateKey, PublicAddress: want}}

	got, err := client.SignerAddress()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("SignerAddress = %s, want %s", got, want)
	}
}

func TestSignerAddressRequiresSigner(t *testing.T) {
	if _, err := (&EthClient{}).SignerAddress(); err == nil {
		t.Fatal("SignerAddress succeeded without a signer")
	}
}

func TestTransferContextHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&EthClient{}).TransferContext(ctx, common.Address{}, common.Address{}, big.NewInt(1))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("TransferContext error = %v, want context.Canceled", err)
	}
}
