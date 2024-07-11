package client

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func generateKeyPair() (pubkey, privkey []byte) {
	key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	pubkey = crypto.S256().Marshal(key.X, key.Y)

	privkey = make([]byte, 32)
	blob := key.D.Bytes()
	copy(privkey[32-len(blob):], blob)

	return pubkey, privkey
}

func TestEthSigner_SignTransaction(t *testing.T) {
	// Generate a new private key for testing
	publickey, privkey := generateKeyPair()

	// unmarsal the public key
	oriPublicKey, err := crypto.UnmarshalPubkey(publickey)

	// Convert the private key to hex
	privateKeyHex := hex.EncodeToString(privkey)

	// Create a new EthSigner instance
	signer, err := NewEthSigner(privateKeyHex)
	assert.NoError(t, err)

	// Create a dummy transaction (in this case, just some random data)
	tx := []byte("dummy transaction data")

	// from tx to hash for signing
	msg := crypto.Keccak256(tx)

	// Sign the transaction
	signature, err := signer.SignTransaction(msg)
	assert.NoError(t, err)

	// Verify the signature
	publicKey, err := crypto.SigToPub(msg, signature)
	assert.NoError(t, err)

	recoveredAddress := crypto.PubkeyToAddress(*publicKey)
	expectedAddress := crypto.PubkeyToAddress(*oriPublicKey)

	assert.Equal(t, expectedAddress.Hex(), recoveredAddress.Hex(), "The recovered address should match the expected address")
}
