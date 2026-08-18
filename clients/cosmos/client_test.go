package cosmos

import (
	"context"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"google.golang.org/grpc"
)

type fakeAccountQuery struct {
	response *authtypes.QueryAccountInfoResponse
}

func (f fakeAccountQuery) AccountInfo(context.Context, *authtypes.QueryAccountInfoRequest, ...grpc.CallOption) (*authtypes.QueryAccountInfoResponse, error) {
	return f.response, nil
}

type fakeBroadcaster struct {
	request  *txservice.BroadcastTxRequest
	response *txservice.BroadcastTxResponse
}

func (f *fakeBroadcaster) BroadcastTx(_ context.Context, request *txservice.BroadcastTxRequest, _ ...grpc.CallOption) (*txservice.BroadcastTxResponse, error) {
	f.request = request
	return f.response, nil
}

func TestTransferBuildsSignsAndBroadcastsMsgSend(t *testing.T) {
	broadcaster := &fakeBroadcaster{response: &txservice.BroadcastTxResponse{
		TxResponse: &sdk.TxResponse{Code: 0, TxHash: "COSMOS_TX_HASH"},
	}}
	client, err := newClient(Config{
		Endpoint: "unused.test:9090",
		Mnemonic: testMnemonic(t),
		ChainID:  "cosmos-test-chain",
		GasLimit: 200000,
		FeeUatom: "5000",
	}, fakeAccountQuery{response: &authtypes.QueryAccountInfoResponse{
		Info: &authtypes.BaseAccount{AccountNumber: 7, Sequence: 9},
	}}, broadcaster)
	if err != nil {
		t.Fatal(err)
	}

	recipientBytes := make([]byte, 20)
	for i := range recipientBytes {
		recipientBytes[i] = byte(i + 1)
	}
	recipient, err := encodeTestAddress(recipientBytes)
	if err != nil {
		t.Fatal(err)
	}
	txHash, err := client.Transfer(context.Background(), recipient, big.NewInt(1234567))
	if err != nil {
		t.Fatal(err)
	}
	if txHash != "COSMOS_TX_HASH" {
		t.Fatalf("txHash = %q", txHash)
	}
	if broadcaster.request == nil || broadcaster.request.Mode != txservice.BroadcastMode_BROADCAST_MODE_SYNC {
		t.Fatalf("broadcast request = %#v", broadcaster.request)
	}

	decoded, err := client.txConfig.TxDecoder()(broadcaster.request.TxBytes)
	if err != nil {
		t.Fatalf("decode transaction: %v", err)
	}
	messages := decoded.GetMsgs()
	if len(messages) != 1 {
		t.Fatalf("message count = %d", len(messages))
	}
	message, ok := messages[0].(*banktypes.MsgSend)
	if !ok {
		t.Fatalf("message type = %T", messages[0])
	}
	if message.FromAddress != client.Address() || message.ToAddress != recipient {
		t.Fatalf("message addresses = %s -> %s", message.FromAddress, message.ToAddress)
	}
	if message.Amount.String() != "1234567uatom" {
		t.Fatalf("message amount = %s", message.Amount)
	}
	feeTx, ok := decoded.(sdk.FeeTx)
	if !ok {
		t.Fatalf("decoded transaction does not implement sdk.FeeTx")
	}
	if feeTx.GetGas() != 200000 || feeTx.GetFee().String() != "5000uatom" {
		t.Fatalf("fee/gas = %s/%d", feeTx.GetFee(), feeTx.GetGas())
	}
	signedTx, ok := decoded.(authsigning.Tx)
	if !ok {
		t.Fatalf("decoded transaction does not implement auth signing Tx")
	}
	signatures, err := signedTx.GetSignaturesV2()
	if err != nil {
		t.Fatalf("read signatures: %v", err)
	}
	if len(signatures) != 1 {
		t.Fatalf("signature count = %d", len(signatures))
	}
	signatureData, ok := signatures[0].Data.(*signingtypes.SingleSignatureData)
	if !ok || len(signatureData.Signature) == 0 {
		t.Fatalf("transaction is not directly signed: %#v", signatures[0].Data)
	}
}

func TestTransferReturnsBroadcastCode(t *testing.T) {
	broadcaster := &fakeBroadcaster{response: &txservice.BroadcastTxResponse{
		TxResponse: &sdk.TxResponse{Code: 12, RawLog: "insufficient fee"},
	}}
	client, err := newClient(Config{
		Endpoint: "unused.test:9090",
		Mnemonic: testMnemonic(t),
		ChainID:  "cosmos-test-chain",
		GasLimit: 1,
		FeeUatom: "1",
	}, fakeAccountQuery{response: &authtypes.QueryAccountInfoResponse{
		Info: &authtypes.BaseAccount{},
	}}, broadcaster)
	if err != nil {
		t.Fatal(err)
	}
	recipient, _ := encodeTestAddress(make([]byte, 20))
	_, err = client.Transfer(context.Background(), recipient, big.NewInt(1))
	if err == nil || err.Error() != "cosmos transaction rejected with code 12: insufficient fee" {
		t.Fatalf("Transfer error = %v", err)
	}
}

func encodeTestAddress(data []byte) (string, error) {
	return addressFromBytes(data)
}

func testMnemonic(t *testing.T) string {
	t.Helper()
	mnemonic, err := bip39.NewMnemonic(make([]byte, 16))
	if err != nil {
		t.Fatalf("create test mnemonic: %v", err)
	}
	return mnemonic
}
