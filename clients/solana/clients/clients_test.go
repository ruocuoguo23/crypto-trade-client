package clients

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSendTransaction(t *testing.T) {
	Convey("Test SendTransaction", t, func() {
		keyPath := "/Users/jeff.wu/.config/solana/id.json"
		client, err := NewSolClient(context.Background(), keyPath)
		So(err, ShouldBeNil)

		// test programID of hello world
		programID := "FyCJ7kDf2RbfoXpuCKT1KhKQxhgbgb9Wj9esDYrm1K6h"
		txHash, err := client.SendTransaction(programID)
		So(err, ShouldBeNil)

		So(txHash, ShouldNotBeEmpty)
		fmt.Println("txHash:", txHash)
	})
}

func TestGetBalance(t *testing.T) {
	Convey("Test GetBalance", t, func() {
		keyPath := "/Users/jeff.wu/.config/solana/id.json"
		client, err := NewSolClient(context.Background(), keyPath)
		So(err, ShouldBeNil)

		balance, err := client.GetBalance()
		So(err, ShouldBeNil)

		So(balance, ShouldBeGreaterThan, 0)
		fmt.Println("balance:", balance)
	})
}
