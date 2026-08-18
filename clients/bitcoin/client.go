package bitcoin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const satoshisPerBitcoin = int64(100_000_000)

type Config struct {
	URL           string
	RPCUser       string
	RPCPassword   string
	Wallet        string
	Network       string
	FromAddress   string
	ConfirmTarget int64
}

type OutPoint struct {
	TxID string
	Vout uint32
}

type Client struct {
	endpoint      string
	rpcUser       string
	rpcPassword   string
	fromAddress   string
	confirmTarget int64
	network       *chaincfg.Params
	httpClient    *http.Client
	requestID     atomic.Uint64
}

func NewClient(config Config) (*Client, error) {
	if config.URL == "" {
		return nil, errors.New("bitcoin RPC URL is required")
	}
	if config.RPCUser == "" || config.RPCPassword == "" {
		return nil, errors.New("bitcoin RPC user and password are required")
	}
	if config.Wallet == "" {
		return nil, errors.New("bitcoin wallet is required")
	}
	if strings.Contains(config.Wallet, "/") {
		return nil, errors.New("bitcoin wallet must not contain a slash")
	}
	if config.FromAddress == "" {
		return nil, errors.New("bitcoin from address is required")
	}

	network, err := NetworkParams(config.Network)
	if err != nil {
		return nil, err
	}
	if err := ValidateAddress(config.FromAddress, network); err != nil {
		return nil, fmt.Errorf("invalid bitcoin from address: %w", err)
	}

	endpoint, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("parse bitcoin RPC URL: %w", err)
	}
	if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return nil, errors.New("bitcoin RPC URL must use http or https")
	}
	endpoint.Path = strings.TrimSuffix(endpoint.Path, "/") + "/wallet/" + url.PathEscape(config.Wallet)

	confirmTarget := config.ConfirmTarget
	if confirmTarget <= 0 {
		confirmTarget = 6
	}

	return &Client{
		endpoint:      endpoint.String(),
		rpcUser:       config.RPCUser,
		rpcPassword:   config.RPCPassword,
		fromAddress:   config.FromAddress,
		confirmTarget: confirmTarget,
		network:       network,
		httpClient:    &http.Client{},
	}, nil
}

func NetworkParams(network string) (*chaincfg.Params, error) {
	switch strings.ToLower(network) {
	case "mainnet":
		return &chaincfg.MainNetParams, nil
	case "testnet", "testnet3":
		return &chaincfg.TestNet3Params, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "signet":
		return &chaincfg.SigNetParams, nil
	default:
		return nil, fmt.Errorf("unsupported bitcoin network %q", network)
	}
}

func ValidateAddress(address string, network *chaincfg.Params) error {
	decoded, err := btcutil.DecodeAddress(address, network)
	if err != nil {
		return err
	}
	if !decoded.IsForNet(network) {
		return fmt.Errorf("address is not for %s", network.Name)
	}
	return nil
}

func ParseOutPoint(value string) (OutPoint, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return OutPoint{}, errors.New("UTXO must use txid:vout format")
	}
	if _, err := chainhash.NewHashFromStr(parts[0]); err != nil {
		return OutPoint{}, fmt.Errorf("invalid UTXO txid: %w", err)
	}
	vout, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return OutPoint{}, fmt.Errorf("invalid UTXO vout: %w", err)
	}
	return OutPoint{TxID: parts[0], Vout: uint32(vout)}, nil
}

func (c *Client) Send(ctx context.Context, toAddress string, satoshis int64, inputs []OutPoint) (string, error) {
	if satoshis <= 0 {
		return "", errors.New("bitcoin amount must be greater than zero")
	}
	if len(inputs) == 0 {
		return "", errors.New("at least one bitcoin UTXO is required")
	}
	if err := ValidateAddress(toAddress, c.network); err != nil {
		return "", fmt.Errorf("invalid bitcoin recipient: %w", err)
	}

	var addressInfo struct {
		IsMine       bool   `json:"ismine"`
		ScriptPubKey string `json:"scriptPubKey"`
	}
	if err := c.call(ctx, "getaddressinfo", []any{c.fromAddress}, &addressInfo); err != nil {
		return "", fmt.Errorf("get bitcoin source address info: %w", err)
	}
	if !addressInfo.IsMine {
		return "", errors.New("bitcoin source address is not owned by the configured wallet")
	}

	seen := make(map[OutPoint]struct{}, len(inputs))
	rpcInputs := make([]map[string]any, 0, len(inputs))
	for _, input := range inputs {
		if _, ok := seen[input]; ok {
			return "", fmt.Errorf("duplicate bitcoin UTXO %s:%d", input.TxID, input.Vout)
		}
		seen[input] = struct{}{}

		var txOut *struct {
			ScriptPubKey struct {
				Hex string `json:"hex"`
			} `json:"scriptPubKey"`
		}
		if err := c.call(ctx, "gettxout", []any{input.TxID, input.Vout, true}, &txOut); err != nil {
			return "", fmt.Errorf("get bitcoin UTXO %s:%d: %w", input.TxID, input.Vout, err)
		}
		if txOut == nil {
			return "", fmt.Errorf("bitcoin UTXO %s:%d is spent or unavailable", input.TxID, input.Vout)
		}
		if txOut.ScriptPubKey.Hex != addressInfo.ScriptPubKey {
			return "", fmt.Errorf("bitcoin UTXO %s:%d does not belong to the configured source address", input.TxID, input.Vout)
		}
		rpcInputs = append(rpcInputs, map[string]any{"txid": input.TxID, "vout": input.Vout})
	}

	outputs := map[string]json.Number{toAddress: formatSatoshis(satoshis)}
	var unsignedHex string
	if err := c.call(ctx, "createrawtransaction", []any{rpcInputs, outputs}, &unsignedHex); err != nil {
		return "", fmt.Errorf("create bitcoin raw transaction: %w", err)
	}

	fundOptions := map[string]any{
		"add_inputs":      false,
		"changeAddress":   c.fromAddress,
		"conf_target":     c.confirmTarget,
		"estimate_mode":   "conservative",
		"lockUnspents":    false,
		"includeWatching": false,
	}
	var funded struct {
		Hex string `json:"hex"`
	}
	if err := c.call(ctx, "fundrawtransaction", []any{unsignedHex, fundOptions}, &funded); err != nil {
		return "", fmt.Errorf("fund bitcoin raw transaction: %w", err)
	}

	var signed struct {
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}
	if err := c.call(ctx, "signrawtransactionwithwallet", []any{funded.Hex}, &signed); err != nil {
		return "", fmt.Errorf("sign bitcoin raw transaction: %w", err)
	}
	if !signed.Complete {
		return "", errors.New("bitcoin wallet did not fully sign the transaction")
	}

	var accepted []struct {
		Allowed      bool   `json:"allowed"`
		RejectReason string `json:"reject-reason"`
	}
	if err := c.call(ctx, "testmempoolaccept", []any{[]string{signed.Hex}}, &accepted); err != nil {
		return "", fmt.Errorf("test bitcoin mempool acceptance: %w", err)
	}
	if len(accepted) != 1 || !accepted[0].Allowed {
		reason := "transaction was rejected"
		if len(accepted) == 1 && accepted[0].RejectReason != "" {
			reason = accepted[0].RejectReason
		}
		return "", fmt.Errorf("bitcoin mempool rejected transaction: %s", reason)
	}

	var txID string
	if err := c.call(ctx, "sendrawtransaction", []any{signed.Hex}, &txID); err != nil {
		return "", fmt.Errorf("broadcast bitcoin transaction: %w", err)
	}
	return txID, nil
}

func formatSatoshis(satoshis int64) json.Number {
	return json.Number(fmt.Sprintf("%d.%08d", satoshis/satoshisPerBitcoin, satoshis%satoshisPerBitcoin))
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("Bitcoin Core RPC error %d: %s", e.Code, e.Message)
}

func (c *Client) call(ctx context.Context, method string, params any, result any) error {
	body, err := json.Marshal(rpcRequest{
		JSONRPC: "1.0",
		ID:      c.requestID.Add(1),
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.rpcUser, c.rpcPassword)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(responseBody, &rpcResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return rpcResp.Error
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(rpcResp.Result, result); err != nil {
		return fmt.Errorf("decode result: %w", err)
	}
	return nil
}
