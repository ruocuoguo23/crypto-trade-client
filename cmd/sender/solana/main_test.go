package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
)

func TestDryRunDoesNotContactSolanaRPC(t *testing.T) {
	wallet := solana.NewWallet()
	keyData, err := json.Marshal([]byte(wallet.PrivateKey))
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(t.TempDir(), "id.json")
	if err := os.WriteFile(keyPath, keyData, 0o600); err != nil {
		t.Fatal(err)
	}
	recipient := solana.NewWallet().PublicKey().String()
	configPath := filepath.Join(t.TempDir(), "chain.yaml")
	configData := fmt.Sprintf(`Chains:
  - name: solana
    url: http://127.0.0.1:1
    keyPath: %s
`, keyPath)
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newCommand(&output)
	command.SetArgs([]string{"--config", configPath, "--to", recipient, "--amount", "0.001"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "mode: dry-run") || !strings.Contains(output.String(), wallet.PublicKey().String()) {
		t.Fatalf("unexpected output:\n%s", output.String())
	}
}
