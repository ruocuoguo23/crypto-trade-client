package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"crypto-trade-client/clients/bitcoin"
	"crypto-trade-client/cmd/sender/internal/amount"
	"crypto-trade-client/cmd/sender/internal/commandutil"

	"github.com/spf13/cobra"
)

func main() {
	command := newCommand(os.Stdout)
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newCommand(output io.Writer) *cobra.Command {
	var configPath string
	var toAddress string
	var amountValue string
	var broadcast bool
	var utxoValues []string

	command := &cobra.Command{
		Use:           "sender-bitcoin",
		Short:         "Send native BTC with selected Bitcoin Core wallet UTXOs",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			chain, err := commandutil.LoadChain(configPath, "bitcoin")
			if err != nil {
				return err
			}
			baseUnits, err := amount.Parse(amountValue, 8)
			if err != nil {
				return fmt.Errorf("invalid BTC amount: %w", err)
			}
			satoshis, err := amount.Int64(baseUnits)
			if err != nil {
				return fmt.Errorf("invalid BTC amount: %w", err)
			}

			inputs := make([]bitcoin.OutPoint, 0, len(utxoValues))
			for _, value := range utxoValues {
				input, err := bitcoin.ParseOutPoint(value)
				if err != nil {
					return fmt.Errorf("invalid --utxo %q: %w", value, err)
				}
				inputs = append(inputs, input)
			}
			if len(inputs) == 0 {
				return fmt.Errorf("at least one --utxo is required")
			}

			client, err := bitcoin.NewClient(bitcoin.Config{
				URL:           chain.URL,
				RPCUser:       chain.RPCUser,
				RPCPassword:   chain.RPCPassword,
				Wallet:        chain.Wallet,
				Network:       chain.Network,
				FromAddress:   chain.FromAddress,
				ConfirmTarget: chain.ConfirmTarget,
			})
			if err != nil {
				return err
			}
			network, err := bitcoin.NetworkParams(chain.Network)
			if err != nil {
				return err
			}
			if err := bitcoin.ValidateAddress(toAddress, network); err != nil {
				return fmt.Errorf("invalid bitcoin recipient: %w", err)
			}

			if !broadcast {
				commandutil.WriteDryRun(output, commandutil.Summary{
					Chain:   "bitcoin",
					From:    chain.FromAddress,
					To:      toAddress,
					Amount:  amountValue + " BTC",
					Network: chain.Network,
					Extra:   []string{fmt.Sprintf("utxos: %d", len(inputs))},
				})
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			txID, err := client.Send(ctx, toAddress, satoshis, inputs)
			if err != nil {
				return err
			}
			fmt.Fprintln(output, txID)
			return nil
		},
	}
	command.Flags().StringVarP(&configPath, "config", "c", "", "path to the ignored YAML configuration file")
	command.Flags().StringVar(&toAddress, "to", "", "recipient Bitcoin address")
	command.Flags().StringVar(&amountValue, "amount", "", "BTC amount as a decimal string")
	command.Flags().StringSliceVar(&utxoValues, "utxo", nil, "wallet UTXO in txid:vout format (repeatable)")
	command.Flags().BoolVar(&broadcast, "broadcast", false, "sign and broadcast the transaction")
	_ = command.MarkFlagRequired("config")
	_ = command.MarkFlagRequired("to")
	_ = command.MarkFlagRequired("amount")
	return command
}
