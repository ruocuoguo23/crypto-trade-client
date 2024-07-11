package client

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestOptimismClient_GetBlock(t *testing.T) {
	Convey("Test GetBlock", t, func() {
		endpoint := "https://practical-green-butterfly.optimism.quiknode.pro/d02f8d49bde8ccbbcec3c9a8962646db998ade83"
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
		endpoint := "https://practical-green-butterfly.optimism.quiknode.pro/d02f8d49bde8ccbbcec3c9a8962646db998ade83"
		client, err := NewOpClient(endpoint)
		So(err, ShouldBeNil)

		height, err := client.GetLatestBlockHeight()
		So(err, ShouldBeNil)
		So(height, ShouldBeGreaterThan, int64(0))
	})
}
