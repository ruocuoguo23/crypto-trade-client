package main

import (
	"crypto-trade-client/clients/polygon"
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
	chain, ok := chainConfig["polygon"]
	if !ok {
		fmt.Println("Core chain not found in chainConfig")
		return
	}

	polyClient, err := polygon.NewPolyClient(chain.URL, chain.PrivateKey)
	if err != nil {
		fmt.Printf("Error creating Core client: %v\n", err)
		return
	}

	// 获取最新区块高度
	latestBlock, err := polyClient.GetLatestBlockHeight()
	if err != nil {
		fmt.Printf("Error getting latest block height: %v\n", err)
		return
	}

	fmt.Printf("Latest block height: %d\n", latestBlock)

	// 开始轮询新区块
	pollNewBlocks(polyClient, latestBlock)
}

func pollNewBlocks(polygonClient *polygon.PolyClient, startHeight int64) {
	currentHeight := startHeight
	for {
		time.Sleep(2 * time.Second)
		block, err := polygonClient.GetBlock("", currentHeight+1)
		if err != nil {
			fmt.Printf("Error retrieving block at height %d: %v\n", currentHeight+1, err)
			continue
		}
		currentHeight++
		fmt.Printf("Retrieved block %d with hash %s\n", currentHeight, block.Hash)
	}
}
