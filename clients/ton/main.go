package main

import (
	"context"
	"encoding/hex"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"log"

	"github.com/xssnick/tonutils-go/ton"
)

func main() {
	// 连接到 TON 节点
	client := liteclient.NewConnectionPool()

	cfg, err := liteclient.GetConfigFromUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		log.Fatalln("get config err: ", err.Error())
		return
	}

	// connect to mainnet lite servers
	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalln("connection err: ", err.Error())
		return
	}

	// initialize ton api lite connection wrapper with full proof checks
	api := ton.NewAPIClient(client, ton.ProofCheckPolicySecure).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	// address on which we are accepting payments
	addr := address.MustParseAddr("UQD4z9HR6C0sSLI001VLFIGqGMzEmRbNE5b_Gr_PuPXi7tzY")

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalln("get masterchain info err: ", err.Error())
		return
	}

	acc, err := api.GetAccount(context.Background(), block, addr)
	if err != nil {
		log.Fatalln("get account err: ", err.Error())
		return
	}

	list, err := api.ListTransactions(context.Background(), addr, 20, acc.LastTxLT, acc.LastTxHash)
	if err != nil {
		log.Fatalln("list transactions err: ", err.Error())
		return
	}

	var hash []byte
	for i := len(list) - 1; i >= 0; i-- {
		ls, err := list[i].IO.Out.ToSlice()
		if err != nil {
			continue
		}

		if len(ls) == 0 {
			continue
		}
		hash = ls[0].Msg.Payload().Hash()
	}

	if hash == nil {
		log.Fatalln("no outs")
	}

	// find tx hash
	tx, err := api.FindLastTransactionByOutMsgHash(context.Background(), addr, hash, 30)
	if err != nil {
		log.Fatalln("cannot find tx:", err.Error())
	}
	log.Printf("tx hash: %s %s\n", hex.EncodeToString(tx.Hash), hex.EncodeToString(acc.LastTxHash))
}
