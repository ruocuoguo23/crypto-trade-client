package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestDryRunDoesNotContactBitcoinRPC(t *testing.T) {
	fromAddress := testBitcoinAddress(t, 1)
	toAddress := testBitcoinAddress(t, 2)
	configPath := filepath.Join(t.TempDir(), "chain.yaml")
	configData := fmt.Sprintf(`Chains:
  - name: bitcoin
    url: http://127.0.0.1:1
    rpcUser: fake-rpc-user
    rpcPassword: fake-rpc-password
    wallet: fake-wallet
    network: regtest
    fromAddress: %s
`, fromAddress)
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newCommand(&output)
	command.SetArgs([]string{
		"--config", configPath,
		"--to", toAddress,
		"--amount", "0.001",
		"--utxo", "0000000000000000000000000000000000000000000000000000000000000001:0",
	})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "mode: dry-run") || !strings.Contains(output.String(), "broadcast: disabled") {
		t.Fatalf("unexpected output:\n%s", output.String())
	}
}

func testBitcoinAddress(t *testing.T, fill byte) string {
	t.Helper()
	address, err := btcutil.NewAddressWitnessPubKeyHash(bytes.Repeat([]byte{fill}, 20), &chaincfg.RegressionNetParams)
	if err != nil {
		t.Fatal(err)
	}
	return address.EncodeAddress()
}
