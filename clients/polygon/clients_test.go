package polygon

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	endpointMainnet = "https://proportionate-spring-brook.matic.quiknode.pro/04711a24cdd335fdbd43663f498e0ba68521823a"
)

func TestPolyClient_GetBlock(t *testing.T) {
	Convey("Test GetBlock", t, func() {
		endpoint := endpointMainnet
		client, err := NewPolyClient(endpoint, "")
		So(err, ShouldBeNil)

		blockHeight := 60019878
		block, err := client.GetBlock("", int64(blockHeight))
		So(err, ShouldBeNil)
		So(block.Number, ShouldEqual, int64(blockHeight))
	})
}

func TestPolyClient_GetLatestBlockHeight(t *testing.T) {
	Convey("Test GetLatestBlockHeight", t, func() {
		endpoint := endpointMainnet
		client, err := NewPolyClient(endpoint, "")
		So(err, ShouldBeNil)

		height, err := client.GetLatestBlockHeight()
		So(err, ShouldBeNil)
		So(height, ShouldBeGreaterThan, int64(0))
	})
}
