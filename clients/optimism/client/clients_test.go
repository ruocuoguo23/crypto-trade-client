package client

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	endpointMainnet = "https://mainnet.optimism.io"
	endpointTestnet = "https://kovan.optimism."
)

func TestOptimismClient_GetBlock(t *testing.T) {
	Convey("Test GetBlock", t, func() {
		endpoint := endpointMainnet
		client, err := NewOpClient(endpoint)
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
		client, err := NewOpClient(endpoint)
		So(err, ShouldBeNil)

		height, err := client.GetLatestBlockHeight()
		So(err, ShouldBeNil)
		So(height, ShouldBeGreaterThan, int64(0))
	})
}
