package ethereum

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Transaction struct {
	innerTx
}

type innerTx struct {
	AccountNonce     string `json:"nonce"`
	GasPrice         string `json:"gasPrice"`
	GasLimit         string `json:"gas"`
	To               string `json:"to"` // nil means contract creation
	Value            string `json:"value"`
	Payload          string `json:"input"`
	Hash             string `json:"hash"`
	BlockNumber      string `json:"blockNumber"`
	BlockHash        string `json:"blockHash,omitempty"`
	From             string `json:"from"`
	TransactionIndex string `json:"transactionIndex"`
}

func (t *Transaction) UnmarshalJSON(bytes []byte) error {
	switch bytes[0] {
	case '"':
		if err := json.Unmarshal(bytes, &t.Hash); err != nil {
			return err
		}
	case '{':
		var itx innerTx
		if err := json.Unmarshal(bytes, &itx); err != nil {
			return err
		}
		t.innerTx = itx
	}
	return nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.innerTx)
}

func (t *Transaction) EncodeToMap() (map[string]interface{}, error) {
	bz, err := t.MarshalJSON()
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{})
	err = json.Unmarshal(bz, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (t *Transaction) DecodeFromMap(m map[string]interface{}) error {
	bz, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return t.UnmarshalJSON(bz)
}

type Receipt struct {
	BlockNumber       string             `json:"blockNumber"`
	BlockHash         string             `json:"blockHash"`
	GasUsed           *hexutil.Big       `json:"gasUsed"`
	EffectiveGasPrice *hexutil.Big       `json:"effectiveGasPrice"`
	Status            hexutil.Uint64     `json:"status"`
	Logs              []ethcoretypes.Log `json:"logs"`
	TransactionHash   string             `json:"transactionHash"`

	// extra fields for optimism
	L1Fee       *hexutil.Big `json:"l1Fee"`
	L1FeeScalar string       `json:"l1FeeScalar"`
	L1GasPrice  *hexutil.Big `json:"l1GasPrice"`
	L1GasUsed   *hexutil.Big `json:"l1GasUsed"`
}

type Block struct {
	Hash       string         `json:"hash"`
	ParentHash string         `json:"parentHash"`
	Difficulty string         `json:"difficulty"` // integer of the difficulty for this block.
	Number     hexutil.Uint64 `json:"number"`     // the block number (height). null when its pending block.
	Time       hexutil.Uint64 `json:"timestamp"`  // the unix timestamp for when the block was collated.
	Size       hexutil.Uint64 `json:"size"`       // integer the size of this block in bytes.
	Nonce      string         `json:"nonce"`      // 8 Bytes - hash of the generated proof-of-work. null when its pending block.

	Transactions []Transaction `json:"transactions"`
}

type Transfer struct {
	ChainId              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int // the maximum fee per gas that we are willing to give to miners (aka: priority fee)
	maxFeePerGas         *big.Int // the maximum fee per gas that we are willing to pay total (aka: max fee), which covers both the priority fee and base fee
	GasLimit             uint64
	To                   []byte
	Value                *big.Int
	Data                 []byte

	// Signature values
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}

type CallFrame struct {
	Type    string      `json:"type"`
	From    string      `json:"from"`
	To      string      `json:"to,omitempty"`
	Value   string      `json:"value,omitempty"`
	Gas     string      `json:"gas"`
	GasUsed string      `json:"gasUsed"`
	Input   string      `json:"input"`
	Output  string      `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
	Calls   []CallFrame `json:"calls,omitempty"`
}

// EvmChainInfo is data element of https://chainid.network/chains.json
type EvmChainInfo struct {
	Name     string   `json:"name"`
	Chain    string   `json:"chain"`
	Icon     string   `json:"icon"`
	Rpc      []string `json:"rpc"`
	Features []struct {
		Name string `json:"name"`
	} `json:"features"`
	NativeCurrency struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"nativeCurrency"`
	InfoURL   string `json:"infoURL"`
	ShortName string `json:"shortName"`
	ChainId   uint64 `json:"chainId"`
	NetworkId uint64 `json:"networkId"`
	Slip44    uint64 `json:"slip44"`
	Explorers []struct {
		Name     string `json:"name"`
		Url      string `json:"url"`
		Standard string `json:"standard"`
	} `json:"explorers"`
}

var chains = []EvmChainInfo{
	{
		Name:      "Ethereum Mainnet",
		Chain:     "ETH",
		Icon:      "ethereum",
		ShortName: "eth",
		ChainId:   1,
	},
	{
		Name:      "Optimism",
		Chain:     "ETH",
		Icon:      "optimism",
		ShortName: "oeth",
		ChainId:   10,
	},
}

func Chains() []EvmChainInfo {
	return chains
}

func ChainMap() map[uint64]EvmChainInfo {
	out := make(map[uint64]EvmChainInfo)
	for _, chain := range chains {
		out[chain.ChainId] = chain
	}
	return out
}
