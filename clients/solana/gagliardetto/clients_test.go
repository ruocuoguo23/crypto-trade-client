package gagliardetto

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

//func TestSendTransaction(t *testing.T) {
//	Convey("Test SendTransaction", t, func() {
//		keyPath := "/Users/jeff.wu/.config/solana/id.json"
//		client, err := NewSolClient(context.Background(), keyPath)
//		So(err, ShouldBeNil)
//
//		recipient := solana.MustPublicKeyFromBase58("这里替换为接收者的公钥")
//		amount := uint64(1000000000) // 1 SOL的lamports数量
//
//		instruction := solana.NewInstruction(
//			client.account.PublicKey(),
//			recipient,
//			amount,
//		).Build()
//
//		txHash, err := client.SendTransaction(instruction)
//		So(err, ShouldBeNil)
//
//		So(txHash, ShouldNotBeEmpty)
//		fmt.Println("txHash:", txHash)
//	})
//}

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
	})
}
