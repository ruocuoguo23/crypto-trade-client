package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	solanaclient "crypto-trade-client/clients/solana/gagliardetto"
	"crypto-trade-client/cmd/sender/internal/amount"
	"crypto-trade-client/cmd/sender/internal/commandutil"

	"github.com/gagliardetto/solana-go"
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

	command := &cobra.Command{
		Use:           "sender-solana",
		Short:         "Send native SOL",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			chain, err := commandutil.LoadChain(configPath, "solana")
			if err != nil {
				return err
			}
			if chain.URL == "" {
				return errors.New("solana RPC URL is required")
			}
			baseUnits, err := amount.Parse(amountValue, 9)
			if err != nil {
				return fmt.Errorf("invalid SOL amount: %w", err)
			}
			lamports, err := amount.Uint64(baseUnits)
			if err != nil {
				return fmt.Errorf("invalid SOL amount: %w", err)
			}
			recipient, err := solana.PublicKeyFromBase58(toAddress)
			if err != nil {
				return fmt.Errorf("invalid solana recipient: %w", err)
			}
			fromAddress, err := solanaclient.AddressFromKeyFile(chain.KeyPath)
			if err != nil {
				return err
			}

			if !broadcast {
				commandutil.WriteDryRun(output, commandutil.Summary{
					Chain:   "solana",
					From:    fromAddress.String(),
					To:      recipient.String(),
					Amount:  amountValue + " SOL",
					Network: "configured RPC",
				})
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := solanaclient.NewSolClientWithEndpoint(ctx, chain.URL, chain.KeyPath)
			if err != nil {
				return err
			}
			signature, err := client.Transfer(toAddress, lamports)
			if err != nil {
				return err
			}
			fmt.Fprintln(output, signature)
			return nil
		},
	}
	command.Flags().StringVarP(&configPath, "config", "c", "", "path to the ignored YAML configuration file")
	command.Flags().StringVar(&toAddress, "to", "", "recipient Solana public key")
	command.Flags().StringVar(&amountValue, "amount", "", "SOL amount as a decimal string")
	command.Flags().BoolVar(&broadcast, "broadcast", false, "sign and broadcast the transaction")
	_ = command.MarkFlagRequired("config")
	_ = command.MarkFlagRequired("to")
	_ = command.MarkFlagRequired("amount")
	return command
}
