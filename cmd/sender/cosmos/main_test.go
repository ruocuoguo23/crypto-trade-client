package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cosmosclient "crypto-trade-client/clients/cosmos"

	"github.com/cosmos/go-bip39"
)

func TestDryRunDoesNotContactCosmosGRPC(t *testing.T) {
	mnemonic := testMnemonic(t, 0)
	recipient, err := cosmosclient.AddressFromMnemonic(testMnemonic(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "chain.yaml")
	configData := fmt.Sprintf(`Chains:
  - name: cosmos
    url: 127.0.0.1:1
    mnemonic: %q
    chainID: theta-testnet-001
    tlsEnabled: false
    gasLimit: 200000
    feeUatom: "5000"
`, mnemonic)
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newCommand(&output)
	command.SetArgs([]string{"--config", configPath, "--to", recipient, "--amount", "0.001"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), mnemonic) || !strings.Contains(output.String(), "mode: dry-run") {
		t.Fatalf("unexpected output:\n%s", output.String())
	}
}

func testMnemonic(t *testing.T, fill byte) string {
	t.Helper()
	mnemonic, err := bip39.NewMnemonic(bytes.Repeat([]byte{fill}, 16))
	if err != nil {
		t.Fatal(err)
	}
	return mnemonic
}
