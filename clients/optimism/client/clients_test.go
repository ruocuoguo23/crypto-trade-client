package client

import (
	. "github.com/smartystreets/goconvey/convey"
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

func TestOptimismClient_GetTransactionCount(t *testing.T) {
	Convey("Test GetTransactionCount", t, func() {
		endpoint := endpointTestnet
		client, err := NewOpClient(endpoint, "")
		So(err, ShouldBeNil)

		// Testnet address for optimism
		addr := "0x9548251949b08521F4397cDfafbB58b50571a2e6"
		nonce, err := client.GetTransactionCount(addr)
		So(err, ShouldBeNil)
		So(nonce, ShouldBeGreaterThan, int64(0))

		// print current nonce
		t.Logf("nonce: %d", nonce)
	})
}
