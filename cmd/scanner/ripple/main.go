package main

import (
	"crypto-trade-client/clients/ripple"
	"crypto-trade-client/common/config"
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

func main() {
	var configPath string

	var rootCmd = &cobra.Command{
		Use:   "scanner",
		Short: "Scanner is a tool for scanning things",
		Run: func(cmd *cobra.Command, args []string) {
			scanner(configPath)
		},
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to the configuration file")
	_ = rootCmd.MarkFlagRequired("config")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func scanner(configPath string) {
	// load chainConfig
	chainConfig, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading chainConfig: %v\n", err)
		return
	}

	// from the chainConfig, get the chain URL
	chain, ok := chainConfig["ripple"]
	if !ok {
		fmt.Println("Ripple chain not found in chainConfig")
		return
	}

	xrpClient, err := ripple.NewXrpClient(chain.URL)
	if err != nil {
		fmt.Printf("Error creating Ripple client: %v\n", err)
		return
	}

	// 获取最新区块高度
	latestBlock, err := xrpClient.LedgerClosed()
	if err != nil {
		fmt.Printf("Error getting latest block height: %v\n", err)
		return
	}

	fmt.Printf("Latest block height: %d\n", latestBlock.Result.LedgerIndex)

	// 开始轮询新区块
	pollNewBlocks(xrpClient, int64(latestBlock.Result.LedgerIndex))
}

func pollNewBlocks(xrpClient *ripple.XrpClient, startHeight int64) {
	currentHeight := startHeight
	for {
		time.Sleep(3 * time.Second)
		block, err := xrpClient.Ledger("", currentHeight+1)
		if err != nil {
			fmt.Printf("Error retrieving block at height %d: %v\n", currentHeight+1, err)
			continue
		}
		currentHeight++
		fmt.Printf("Retrieved block %d with hash %s\n", currentHeight, block.Result.LedgerHash)

		// get the transactions in the block
	}
}