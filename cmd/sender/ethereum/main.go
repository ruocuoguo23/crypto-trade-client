package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	ethclient "crypto-trade-client/clients/ethereum/client"
	"crypto-trade-client/cmd/sender/internal/amount"
	"crypto-trade-client/cmd/sender/internal/commandutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
		Use:           "sender-ethereum",
		Short:         "Send native ETH",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			chain, err := commandutil.LoadChain(configPath, "ethereum")
			if err != nil {
				return err
			}
			if chain.URL == "" {
				return errors.New("ethereum RPC URL is required")
			}
			wei, err := amount.Parse(amountValue, 18)
			if err != nil {
				return fmt.Errorf("invalid ETH amount: %w", err)
			}
			if !common.IsHexAddress(toAddress) {
				return errors.New("invalid ethereum recipient address")
			}
			expectedChainID := new(big.Int)
			if _, ok := expectedChainID.SetString(chain.ChainID, 10); !ok || expectedChainID.Sign() <= 0 {
				return errors.New("ethereum chainID must be a positive decimal integer")
			}
			privateKeyHex := strings.TrimPrefix(chain.PrivateKey, "0x")
			privateKey, err := crypto.HexToECDSA(privateKeyHex)
			if err != nil {
				return fmt.Errorf("parse ethereum private key: %w", err)
			}
			fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

			if !broadcast {
				commandutil.WriteDryRun(output, commandutil.Summary{
					Chain:   "ethereum",
					From:    fromAddress.Hex(),
					To:      common.HexToAddress(toAddress).Hex(),
					Amount:  amountValue + " ETH",
					Network: "chain-id " + expectedChainID.String(),
				})
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := ethclient.NewEthClient(chain.URL, "ethereum", privateKeyHex)
			if err != nil {
				return fmt.Errorf("create ethereum client: %w", err)
			}
			remoteChainID, err := client.GetChainID(ctx)
			if err != nil {
				return fmt.Errorf("query ethereum chain ID: %w", err)
			}
			if err := validateChainID(expectedChainID, remoteChainID); err != nil {
				return err
			}
			txHash, err := client.TransferContext(ctx, fromAddress, common.HexToAddress(toAddress), wei)
			if err != nil {
				return err
			}
			fmt.Fprintln(output, txHash.Hex())
			return nil
		},
	}
	command.Flags().StringVarP(&configPath, "config", "c", "", "path to the ignored YAML configuration file")
	command.Flags().StringVar(&toAddress, "to", "", "recipient Ethereum address")
	command.Flags().StringVar(&amountValue, "amount", "", "ETH amount as a decimal string")
	command.Flags().BoolVar(&broadcast, "broadcast", false, "sign and broadcast the transaction")
	_ = command.MarkFlagRequired("config")
	_ = command.MarkFlagRequired("to")
	_ = command.MarkFlagRequired("amount")
	return command
}

func validateChainID(expected, actual *big.Int) error {
	if expected == nil || actual == nil || expected.Cmp(actual) != 0 {
		return fmt.Errorf("ethereum chain ID mismatch: RPC returned %v, config requires %v", actual, expected)
	}
	return nil
}
