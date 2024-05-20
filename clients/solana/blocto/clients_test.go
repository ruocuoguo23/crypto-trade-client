package blocto

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

		// test programID of hello world on localnet
		//programID := "FyCJ7kDf2RbfoXpuCKT1KhKQxhgbgb9Wj9esDYrm1K6h"

		// test programID of hello world on devnet
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

func TestGetTransaction(t *testing.T) {
	Convey("Test GetTransaction", t, func() {
		keyPath := "/Users/jeff.wu/.config/solana/id.json"
		client, err := NewSolClient(context.Background(), keyPath)
		So(err, ShouldBeNil)

		// txHash of hello world on devnet
		txHash := "3DxqoU9JvXje49zVZZYSwD1f8hGyZaikP1o3ZPNmgCrrKai1yFZ5oXQbzQMN7p8QTFzKdAaiU9uYGm1fX4G1xkAz"
		transaction, err := client.GetTransaction(txHash)
		So(err, ShouldBeNil)

		So(transaction, ShouldNotBeNil)
		fmt.Println("transaction:", transaction)

		// print transaction log for debugging
		for _, message := range transaction.Meta.LogMessages {
			fmt.Println("message: ", message)
		}

		// print transaction fee for debugging
		fmt.Println("fee:", transaction.Meta.Fee)
	})
}
