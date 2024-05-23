package ethereum

import (
	"crypto-trade-client/common/rpc"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
)

// EthRpc is eth namespace rpc endpoints
type EthRpc struct {
	ProtocolVersion       func() (string, error)
	BlockNumber           func() (hexutil.Uint64, error)
	GetBlockByNumber      func(EthBlockNumArg, bool) (Block, error)
	GetBlockByHash        func(string, bool) (Block, error)
	GetTransactionReceipt func(string) (Receipt, error)
	GetLogs               func(FilterOption) ([]ethcoretypes.Log, error)
	GetBalance            func(string, EthBlockNumArg) (*hexutil.Big, error)
	Call                  func(CallOption, EthBlockNumArg) (interface{}, error)
	GasPrice              func() (*hexutil.Big, error) // after London will return the exact same number based on the total fees paid (tip + base)
	MaxPriorityFeePerGas  func() (*hexutil.Big, error) // geth only, eth_maxPriorityFeePerGas after London will effectively return eth_gasPrice - baseFee
	GetTransactionCount   func(string, EthBlockNumArg) (hexutil.Uint64, error)
	SendRawTransaction    func(hexutil.Bytes) (hexutil.Bytes, error)
	EstimateGas           func(CallOption, EthBlockNumArg) (hexutil.Uint64, error)
	ChainId               func() (hexutil.Uint64, error)
}

func (r *EthRpc) MethodNamingConvention() rpc.NamingConvention {
	return rpc.CamelCase
}

func (r *EthRpc) Namespace() string {
	return "eth"
}

func (r *EthRpc) NamespaceSeparator() string {
	return "_"
}

type FilterOption struct {
	FromBlock EthBlockNumArg `json:"fromBlock,omitempty"`
	ToBlock   EthBlockNumArg `json:"toBlock,omitempty"`
	Address   []string       `json:"address,omitempty"`
	Topics    []interface{}  `json:"topics,omitempty"`
}

type CallOption struct {
	From  string         `json:"from,omitempty"`
	To    string         `json:"to"` // 20 Bytes - The address the transaction is directed to.
	Value hexutil.Uint64 `json:"value,omitempty"`
	Data  string         `json:"data"` // (optional) Hash of the method signature and encoded parameters. For details see Ethereum ContractAddr ABI in the Solidity documentation
}

type EthBlockNumArg int64

const (
	Latest  EthBlockNumArg = 0
	Pending EthBlockNumArg = -1
)

func (arg EthBlockNumArg) MarshalJSON() ([]byte, error) {
	if arg == Latest {
		return json.Marshal("latest")
	}
	if arg == Pending {
		return json.Marshal("pending")
	}
	if arg < -1 {
		return nil, fmt.Errorf("incorrect block number argument %v", arg)
	}

	return json.Marshal(hexutil.EncodeUint64(uint64(arg)))
}

func (arg *EthBlockNumArg) UnmarshalJSON(bytes []byte) error {
	data := string(bytes)
	if data == "latest" {
		*arg = Latest
	}
	if data == "pending" {
		*arg = Pending
	}
	v, err := hexutil.DecodeUint64(data)
	if err != nil {
		return err
	}
	*arg = EthBlockNumArg(v)

	return nil
}

// EthDebugRpc is debug namespace rpc endpoints for go-ethereum
type EthDebugRpc struct {
	TraceTransaction func(txid string, option TraceOption) (CallFrame, error)
}

func (r *EthDebugRpc) MethodNamingConvention() rpc.NamingConvention {
	return rpc.CamelCase
}

func (r *EthDebugRpc) Namespace() string {
	return "debug"
}

func (r *EthDebugRpc) NamespaceSeparator() string {
	return "_"
}

type TraceOption struct {
	DisableStorage   bool   `json:"disableStorage"`   // Setting this to true will disable storage capture (default = false).
	DisableStack     bool   `json:"disableStack"`     // Setting this to true will disable stack capture (default = false).
	EnableMemory     bool   `json:"enableMemory"`     // Setting this to true will enable memory capture (default = false).
	EnableReturnData bool   `json:"enableReturnData"` // Setting this to true will enable return data capture (default = false).
	Tracer           string `json:"tracer"`           // Setting this will enable JavaScript-based transaction tracing, described below. If set, the previous four arguments will be ignored.
}
