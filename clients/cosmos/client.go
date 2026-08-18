package cosmos

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"

	cosmosmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	AddressPrefix = "cosmos"
	Denom         = "uatom"
)

type Config struct {
	Endpoint   string
	Mnemonic   string
	ChainID    string
	TLSEnabled bool
	GasLimit   uint64
	FeeUatom   string
}

type accountQuerier interface {
	AccountInfo(context.Context, *authtypes.QueryAccountInfoRequest, ...grpc.CallOption) (*authtypes.QueryAccountInfoResponse, error)
}

type txBroadcaster interface {
	BroadcastTx(context.Context, *txservice.BroadcastTxRequest, ...grpc.CallOption) (*txservice.BroadcastTxResponse, error)
}

type Client struct {
	address      string
	privateKey   cryptotypes.PrivKey
	chainID      string
	gasLimit     uint64
	feeUatom     cosmosmath.Int
	txConfig     client.TxConfig
	accountQuery accountQuerier
	broadcaster  txBroadcaster
	connection   *grpc.ClientConn
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	if config.TLSEnabled {
		dialOption = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}))
	}
	connection, err := grpc.DialContext(ctx, config.Endpoint, dialOption)
	if err != nil {
		return nil, fmt.Errorf("connect to cosmos gRPC endpoint: %w", err)
	}

	cosmosClient, err := newClient(config, authtypes.NewQueryClient(connection), txservice.NewServiceClient(connection))
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	cosmosClient.connection = connection
	return cosmosClient, nil
}

func newClient(config Config, accountQuery accountQuerier, broadcaster txBroadcaster) (*Client, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	if accountQuery == nil || broadcaster == nil {
		return nil, errors.New("cosmos query and broadcast services are required")
	}

	feeUatom, _ := cosmosmath.NewIntFromString(config.FeeUatom)
	privateKeyBytes, err := hd.Secp256k1.Derive()(config.Mnemonic, "", sdk.FullFundraiserPath)
	if err != nil {
		return nil, fmt.Errorf("derive cosmos private key: %w", err)
	}
	privateKey := hd.Secp256k1.Generate()(privateKeyBytes)
	address, err := addressFromBytes(privateKey.PubKey().Address())
	if err != nil {
		return nil, fmt.Errorf("encode cosmos address: %w", err)
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(protoCodec, []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_DIRECT,
	})

	return &Client{
		address:      address,
		privateKey:   privateKey,
		chainID:      config.ChainID,
		gasLimit:     config.GasLimit,
		feeUatom:     feeUatom,
		txConfig:     txConfig,
		accountQuery: accountQuery,
		broadcaster:  broadcaster,
	}, nil
}

func ValidateConfig(config Config) error {
	if config.Endpoint == "" {
		return errors.New("cosmos gRPC endpoint is required")
	}
	if config.Mnemonic == "" {
		return errors.New("cosmos mnemonic is required")
	}
	if config.ChainID == "" {
		return errors.New("cosmos chain ID is required")
	}
	if config.GasLimit == 0 {
		return errors.New("cosmos gas limit must be greater than zero")
	}
	feeUatom, ok := cosmosmath.NewIntFromString(config.FeeUatom)
	if !ok || !feeUatom.IsPositive() {
		return errors.New("cosmos feeUatom must be a positive integer")
	}
	return nil
}

func AddressFromMnemonic(mnemonic string) (string, error) {
	privateKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", sdk.FullFundraiserPath)
	if err != nil {
		return "", fmt.Errorf("derive cosmos private key: %w", err)
	}
	privateKey := hd.Secp256k1.Generate()(privateKeyBytes)
	address, err := addressFromBytes(privateKey.PubKey().Address())
	if err != nil {
		return "", fmt.Errorf("encode cosmos address: %w", err)
	}
	return address, nil
}

func addressFromBytes(data []byte) (string, error) {
	return bech32.ConvertAndEncode(AddressPrefix, data)
}

func ValidateAddress(address string) error {
	prefix, data, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return err
	}
	if prefix != AddressPrefix {
		return fmt.Errorf("address prefix must be %q", AddressPrefix)
	}
	if len(data) != 20 {
		return fmt.Errorf("cosmos address must contain 20 bytes, got %d", len(data))
	}
	return nil
}

func (c *Client) Address() string {
	return c.address
}

func (c *Client) Close() error {
	if c.connection == nil {
		return nil
	}
	return c.connection.Close()
}

func (c *Client) Transfer(ctx context.Context, toAddress string, amount *big.Int) (string, error) {
	if amount == nil || amount.Sign() <= 0 {
		return "", errors.New("cosmos amount must be greater than zero")
	}
	if err := ValidateAddress(toAddress); err != nil {
		return "", fmt.Errorf("invalid cosmos recipient: %w", err)
	}

	accountResponse, err := c.accountQuery.AccountInfo(ctx, &authtypes.QueryAccountInfoRequest{Address: c.address})
	if err != nil {
		return "", fmt.Errorf("query cosmos account: %w", err)
	}
	if accountResponse == nil || accountResponse.Info == nil {
		return "", errors.New("cosmos account query returned no account information")
	}

	message := &banktypes.MsgSend{
		FromAddress: c.address,
		ToAddress:   toAddress,
		Amount: sdk.NewCoins(sdk.NewCoin(
			Denom,
			cosmosmath.NewIntFromBigInt(new(big.Int).Set(amount)),
		)),
	}
	txBuilder := c.txConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(message); err != nil {
		return "", fmt.Errorf("set cosmos transaction message: %w", err)
	}
	txBuilder.SetGasLimit(c.gasLimit)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(Denom, c.feeUatom)))

	emptySignature := signingtypes.SignatureV2{
		PubKey: c.privateKey.PubKey(),
		Data: &signingtypes.SingleSignatureData{
			SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
		},
		Sequence: accountResponse.Info.Sequence,
	}
	if err := txBuilder.SetSignatures(emptySignature); err != nil {
		return "", fmt.Errorf("set cosmos signer information: %w", err)
	}

	signerData := authsigning.SignerData{
		Address:       c.address,
		ChainID:       c.chainID,
		AccountNumber: accountResponse.Info.AccountNumber,
		Sequence:      accountResponse.Info.Sequence,
		PubKey:        c.privateKey.PubKey(),
	}
	signature, err := clienttx.SignWithPrivKey(
		ctx,
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder,
		c.privateKey,
		c.txConfig,
		accountResponse.Info.Sequence,
	)
	if err != nil {
		return "", fmt.Errorf("sign cosmos transaction: %w", err)
	}
	if err := txBuilder.SetSignatures(signature); err != nil {
		return "", fmt.Errorf("set cosmos transaction signature: %w", err)
	}

	txBytes, err := c.txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", fmt.Errorf("encode cosmos transaction: %w", err)
	}
	broadcastResponse, err := c.broadcaster.BroadcastTx(ctx, &txservice.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txservice.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	if err != nil {
		return "", fmt.Errorf("broadcast cosmos transaction: %w", err)
	}
	if broadcastResponse == nil || broadcastResponse.TxResponse == nil {
		return "", errors.New("cosmos broadcast returned no transaction response")
	}
	if broadcastResponse.TxResponse.Code != 0 {
		return "", fmt.Errorf(
			"cosmos transaction rejected with code %d: %s",
			broadcastResponse.TxResponse.Code,
			broadcastResponse.TxResponse.RawLog,
		)
	}
	return broadcastResponse.TxResponse.TxHash, nil
}
