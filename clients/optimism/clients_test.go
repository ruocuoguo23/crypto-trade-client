package optimism

import (
	"context"
	"crypto-trade-client/common/config"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"testing"
)

var (
	endpointMainnet = "https://mainnet.optimism.io"
	endpointTestnet = "https://optimism-sepolia.infura.io/v3/559af310b68646d8accf0cf36111f2eb"
)

func TestOptimismClient_GetBlock(t *testing.T) {
	Convey("Test GetBlock", t, func() {
		endpoint := endpointMainnet
		client, err := NewOpClient(endpoint, "")
		So(err, ShouldBeNil)

		blockHeight := 120167305
		block, err := client.GetBlock("", int64(blockHeight))
		So(err, ShouldBeNil)
		So(block.Number, ShouldEqual, int64(blockHeight))
	})
}

func TestOptimismClient_GetLatestBlockHeight(t *testing.T) {
	Convey("Test GetLatestBlockHeight", t, func() {
		endpoint := endpointMainnet
		client, err := NewOpClient(endpoint, "")
		So(err, ShouldBeNil)

		height, err := client.GetLatestBlockHeight()
		So(err, ShouldBeNil)
		So(height, ShouldBeGreaterThan, int64(0))
	})
}

func TestOptimismClient_PendingNonceAt(t *testing.T) {
	Convey("Test GetTransactionCount", t, func() {
		endpoint := endpointTestnet
		client, err := NewOpClient(endpoint, "")
		So(err, ShouldBeNil)

		// Testnet address for optimism
		addr := "0x9548251949b08521F4397cDfafbB58b50571a2e6"
		ctx := context.Background()
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(addr))
		So(err, ShouldBeNil)
		So(nonce, ShouldBeGreaterThan, int64(0))

		// print current nonce
		t.Logf("nonce: %d", nonce)
	})
}

func TestOptimismClient_Transfer(t *testing.T) {
	Convey("Test Transfer", t, func() {
		configPath := "../../../configs/chain.yaml"
		chainConfig, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Error loading chainConfig: %v", err)
			return
		}

		chain, ok := chainConfig["optimism"]
		if !ok {
			t.Fatalf("Optimism chain not found in chainConfig")
			return
		}

		// Path to the configuration file
		endpoint := endpointTestnet
		client, err := NewOpClient(endpoint, chain.PrivateKey)
		So(err, ShouldBeNil)

		// Testnet addresses for optimism
		signer := common.HexToAddress("0x9548251949b08521F4397cDfafbB58b50571a2e6")
		to := common.HexToAddress("0xebdBa70B23edf9A69B7872b5Ff9A6Ba55e2F2FE4")
		value := big.NewInt(100000000000000) // 0.0001 ETH in wei

		txHash, err := client.Transfer(signer, to, value)
		So(err, ShouldBeNil)

		// Print transaction hash
		t.Logf("Transaction hash: %s", txHash.Hex())

		// 0x86b422e9611adacde5c7f06ce48d60ac9d3dd1c448541fdc31ad887aaba8ede3
	})
}
