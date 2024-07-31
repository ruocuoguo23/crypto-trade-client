package client

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	endpointMainnet = "https://rpc.ankr.com/core/d81edae614e6ff96f295baf03da9276f697e82c871a2af207bf4644d06a7c437"
)

func TestCoreClient_GetBlock(t *testing.T) {
	Convey("Test GetBlock", t, func() {
		endpoint := endpointMainnet
		client, err := NewCoreClient(endpoint, "")
		So(err, ShouldBeNil)

		blockHeight := 16364779
		block, err := client.GetBlock("", int64(blockHeight))
		So(err, ShouldBeNil)
		So(block.Number, ShouldEqual, int64(blockHeight))
	})
}

func TestCoreClient_GetLatestBlockHeight(t *testing.T) {
	Convey("Test GetLatestBlockHeight", t, func() {
		endpoint := endpointMainnet
		client, err := NewCoreClient(endpoint, "")
		So(err, ShouldBeNil)

		height, err := client.GetLatestBlockHeight()
		So(err, ShouldBeNil)
		So(height, ShouldBeGreaterThan, int64(0))
	})
}
