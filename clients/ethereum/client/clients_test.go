package client

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEthClient_GetBlock(t *testing.T) {
	Convey("Test GetBlock", t, func() {
		endpoint := "https://practical-green-butterfly.optimism.quiknode.pro/d02f8d49bde8ccbbcec3c9a8962646db998ade83"
		client, err := NewEthClient(endpoint, "optimism")
		So(err, ShouldBeNil)

		bh, err := client.bestBlockHeader()
		So(err, ShouldBeNil)
		So(bh.Height, ShouldBeGreaterThan, 0)
	})
}
