package commandutil

import (
	"errors"
	"fmt"
	"io"

	"crypto-trade-client/common/config"
)

func LoadChain(configPath string, chainName string) (config.Chain, error) {
	if configPath == "" {
		return config.Chain{}, errors.New("config path is required")
	}
	chains, err := config.LoadConfig(configPath)
	if err != nil {
		return config.Chain{}, err
	}
	chain, ok := chains[chainName]
	if !ok {
		return config.Chain{}, fmt.Errorf("chain %q was not found in config", chainName)
	}
	return chain, nil
}

type Summary struct {
	Chain   string
	From    string
	To      string
	Amount  string
	Network string
	Extra   []string
}

func WriteDryRun(w io.Writer, summary Summary) {
	fmt.Fprintf(w, "chain: %s\n", summary.Chain)
	fmt.Fprintln(w, "mode: dry-run")
	fmt.Fprintf(w, "from: %s\n", summary.From)
	fmt.Fprintf(w, "to: %s\n", summary.To)
	fmt.Fprintf(w, "amount: %s\n", summary.Amount)
	if summary.Network != "" {
		fmt.Fprintf(w, "network: %s\n", summary.Network)
	}
	for _, line := range summary.Extra {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w, "broadcast: disabled (use --broadcast to send)")
}
