package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestDryRunDoesNotContactEthereumRPC(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	recipientKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	recipient := crypto.PubkeyToAddress(recipientKey.PublicKey).Hex()

	configPath := filepath.Join(t.TempDir(), "chain.yaml")
	configData := fmt.Sprintf(`Chains:
  - name: ethereum
    url: http://127.0.0.1:1
    privateKey: %s
    chainID: "11155111"
`, privateKeyHex)
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newCommand(&output)
	command.SetArgs([]string{"--config", configPath, "--to", recipient, "--amount", "0.001"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), privateKeyHex) || !strings.Contains(output.String(), "mode: dry-run") {
		t.Fatalf("unexpected output:\n%s", output.String())
	}
}

func TestValidateChainIDRejectsMismatch(t *testing.T) {
	err := validateChainID(big.NewInt(11155111), big.NewInt(1))
	if err == nil || !strings.Contains(err.Error(), "chain ID mismatch") {
		t.Fatalf("validateChainID error = %v", err)
	}
	if err := validateChainID(big.NewInt(1), big.NewInt(1)); err != nil {
		t.Fatalf("validateChainID matching IDs: %v", err)
	}
	if err := validateChainID(big.NewInt(1), nil); err == nil {
		t.Fatal("validateChainID accepted a nil RPC chain ID")
	}
}
