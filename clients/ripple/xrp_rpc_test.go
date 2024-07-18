package ripple

import (
	"github.com/hashicorp/go-hclog"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	rpcEndpoint = "https://s2.ripple.com:51234/"
	logger      = hclog.New(&hclog.LoggerOptions{
		Name:  "test",
		Level: hclog.LevelFromString("DEBUG"),
	})
)

func TestXrpRpc_LedgerClosed(t *testing.T) {
	Convey("Test LedgerClosed", t, func() {
		client, err := NewXrpRpc(rpcEndpoint, logger)
		So(err, ShouldBeNil)

		resp, err := client.LedgerClosed()
		So(err, ShouldBeNil)
		So(resp.Result.Status, ShouldEqual, "success")
	})
}

func TestXrpRpc_Ledger(t *testing.T) {
	Convey("Test Ledger", t, func() {
		client, err := NewXrpRpc(rpcEndpoint, logger)
		So(err, ShouldBeNil)

		hash := "630FD835CD15D35418699CD80DFF0A52EBE585D676F78ECCA34A02ACB2889CCC"
		height := int64(89443608)
		resp, err := client.Ledger(hash, height)
		So(err, ShouldBeNil)
		So(resp.Result.Status, ShouldEqual, "success")
	})
}

func TestXrpRpc_Tx(t *testing.T) {
	Convey("Test Tx", t, func() {
		client, err := NewXrpRpc(rpcEndpoint, logger)
		So(err, ShouldBeNil)

		hash := "562259FD65583C544B2B328EF3F51007E20620C90746DA32E17CC0306C0CB812"
		resp, err := client.Tx(hash)
		So(err, ShouldBeNil)
		So(resp.Result.Status, ShouldEqual, "success")
	})
}

func TestXrpRpc_Fee(t *testing.T) {
	Convey("Test Fee", t, func() {
		client, err := NewXrpRpc(rpcEndpoint, logger)
		So(err, ShouldBeNil)

		resp, err := client.Fee()
		So(err, ShouldBeNil)
		So(resp.Result.Status, ShouldEqual, "success")
	})
}
