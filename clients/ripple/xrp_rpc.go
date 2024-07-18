package ripple

import (
	"crypto-trade-client/common/web/fetch"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
)

type XrpRpc struct {
	logger hclog.Logger
	client *fetch.Client
}

func NewXrpRpc(rpcEndpoint string, l hclog.Logger) (*XrpRpc, error) {
	return &XrpRpc{
		logger: l,
		client: fetch.NewClientWithEndpoint(rpcEndpoint, l),
	}, nil
}

func (r *XrpRpc) LedgerClosed() (*LedgerClosedResp, error) {
	resp, err := r.client.Post("").
		SetHeaders(map[string]string{"Content-Type": "application/json"}).
		SetBody(map[string]interface{}{
			"id":     uuid.New(),
			"method": "ledger_closed",
		}).Execute()
	if err != nil {
		return nil, err
	}
	var p LedgerClosedResp
	err = json.Unmarshal(resp.BodyBytes(), &p)
	if err != nil {
		return nil, err
	}
	if p.Result.Status != "success" {
		r.logger.Error("response is error for ledger_closed", "resp", string(resp.BodyBytes()))
		return nil, errors.New("get ledger closed failed. json-rpc response with error msg")
	}

	return &p, nil
}

func (r *XrpRpc) Ledger(hash string, height int64) (*LedgerResp, error) {
	resp, err := r.client.Post("").
		SetHeaders(map[string]string{"Content-Type": "application/json"}).
		SetBody(map[string]interface{}{
			"method": "ledger",
			"params": []map[string]interface{}{
				{
					"id":           uuid.New(),
					"ledger_hash":  hash,
					"ledger_index": height,
					"transactions": true,
					// do not use binary for ledger information like close time will be encoded as binary
					//"binary": true,
				},
			},
		}).Execute()
	if err != nil {
		return nil, err
	}
	var ledger LedgerResp
	err = json.Unmarshal(resp.BodyBytes(), &ledger)
	if err != nil {
		return nil, err
	}
	if ledger.Result.Status != "success" {
		r.logger.Error("response is not valid for ledger", "resp", string(resp.BodyBytes()))
		return nil, errors.New("get ledger failed. json-rpc response with error msg")
	}
	return &ledger, nil
}

func (r *XrpRpc) Tx(hash string) (*TxResp, error) {
	resp, err := r.client.Post("").
		SetHeaders(map[string]string{"Content-Type": "application/json"}).
		SetBody(map[string]interface{}{
			"method": "tx",
			"params": []map[string]interface{}{
				{
					"transaction": hash,
				},
			},
		}).Execute()
	if err != nil {
		return nil, err
	}
	var tx TxResp
	err = json.Unmarshal(resp.BodyBytes(), &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *XrpRpc) Fee() (*FeeResp, error) {
	resp, err := r.client.Post("").SetBody(map[string]interface{}{
		"method": "fee",
		"params": []map[string]interface{}{},
	}).SetHeaders(map[string]string{"Content-Type": "application/json"}).Execute()
	if err != nil {
		return nil, err
	}
	var p FeeResp
	err = json.Unmarshal(resp.BodyBytes(), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
