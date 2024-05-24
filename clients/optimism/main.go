package main

import (
	"crypto-trade-client/clients/optimism/client"
	"fmt"
	"time"
)

func main() {
	endpoint := "https://practical-green-butterfly.optimism.quiknode.pro/d02f8d49bde8ccbbcec3c9a8962646db998ade83"
	ethClient, err := client.NewOpClient(endpoint)
	if err != nil {
		fmt.Printf("Error creating Ethereum client: %v\n", err)
		return
	}

	// 获取最新区块高度
	latestBlock, err := ethClient.GetLatestBlockHeight()
	if err != nil {
		fmt.Printf("Error getting latest block height: %v\n", err)
		return
	}

	fmt.Printf("Latest block height: %d\n", latestBlock)

	// 开始轮询新区块
	pollNewBlocks(ethClient, latestBlock)
}

func pollNewBlocks(optimismClient *client.OptimismClient, startHeight int64) {
	currentHeight := startHeight
	for {
		time.Sleep(1 * time.Second)
		block, err := optimismClient.GetBlock("", currentHeight+1)
		if err != nil {
			fmt.Printf("Error retrieving block at height %d: %v\n", currentHeight+1, err)
			continue
		}
		currentHeight++
		fmt.Printf("Retrieved block %d with hash %s\n", currentHeight, block.Hash)
	}
}
