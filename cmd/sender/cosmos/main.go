package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	cosmosclient "crypto-trade-client/clients/cosmos"
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

	command := &cobra.Command{
		Use:           "sender-cosmos",
		Short:         "Send native ATOM on Cosmos Hub",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			chain, err := commandutil.LoadChain(configPath, "cosmos")
			if err != nil {
				return err
			}
			clientConfig := cosmosclient.Config{
				Endpoint:   chain.URL,
				Mnemonic:   chain.Mnemonic,
				ChainID:    chain.ChainID,
				TLSEnabled: chain.TLSEnabled,
				GasLimit:   chain.GasLimit,
				FeeUatom:   chain.FeeUatom,
			}
			if err := cosmosclient.ValidateConfig(clientConfig); err != nil {
				return err
			}
			uatom, err := amount.Parse(amountValue, 6)
			if err != nil {
				return fmt.Errorf("invalid ATOM amount: %w", err)
			}
			if err := cosmosclient.ValidateAddress(toAddress); err != nil {
				return fmt.Errorf("invalid cosmos recipient: %w", err)
			}
			fromAddress, err := cosmosclient.AddressFromMnemonic(chain.Mnemonic)
			if err != nil {
				return err
			}

			if !broadcast {
				commandutil.WriteDryRun(output, commandutil.Summary{
					Chain:   "cosmos",
					From:    fromAddress,
					To:      toAddress,
					Amount:  amountValue + " ATOM",
					Network: chain.ChainID,
					Extra: []string{
						fmt.Sprintf("gas-limit: %d", chain.GasLimit),
						"fee: " + chain.FeeUatom + " uatom",
					},
				})
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := cosmosclient.NewClient(ctx, clientConfig)
			if err != nil {
				return err
			}
			defer client.Close()
			txHash, err := client.Transfer(ctx, toAddress, uatom)
			if err != nil {
				return err
			}
			fmt.Fprintln(output, txHash)
			return nil
		},
	}
	command.Flags().StringVarP(&configPath, "config", "c", "", "path to the ignored YAML configuration file")
	command.Flags().StringVar(&toAddress, "to", "", "recipient Cosmos Hub address")
	command.Flags().StringVar(&amountValue, "amount", "", "ATOM amount as a decimal string")
	command.Flags().BoolVar(&broadcast, "broadcast", false, "sign and broadcast the transaction")
	_ = command.MarkFlagRequired("config")
	_ = command.MarkFlagRequired("to")
	_ = command.MarkFlagRequired("amount")
	return command
}
