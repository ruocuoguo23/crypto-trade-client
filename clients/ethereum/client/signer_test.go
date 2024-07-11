package client

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestEthSigner_SignTransaction(t *testing.T) {
	// Generate a new private key for testing
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	assert.NoError(t, err)

	// Convert the private key to a hex string
	privateKeyHex := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Create a new EthSigner instance
	signer, err := NewEthSigner(privateKeyHex)
	assert.NoError(t, err)

	// Create a dummy transaction (in this case, just some random data)
	tx := []byte("dummy transaction data")

	// Sign the transaction
	signature, err := signer.SignTransaction(tx)
	assert.NoError(t, err)

	// Verify the signature
	publicKey, err := crypto.SigToPub(crypto.Keccak256(tx), signature)
	assert.NoError(t, err)

	recoveredAddress := crypto.PubkeyToAddress(*publicKey)
	expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	assert.Equal(t, expectedAddress.Hex(), recoveredAddress.Hex(), "The recovered address should match the expected address")
}
