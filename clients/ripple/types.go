package ripple

import "github.com/rubblelabs/ripple/data"

type LedgerClosedResp struct {
	Result struct {
		LedgerHash  string `json:"ledger_hash"`
		LedgerIndex int    `json:"Ledger_index"`
		Status      string `json:"status"`
	} `json:"result"`
}

type LedgerResp struct {
	Result struct {
		Ledger struct {
			Closed bool `json:"closed"`
			// The time this ledger was closed, in seconds since the Ripple Epoch.
			// This number measures the number of seconds since the "Ripple Epoch" of January 1, 2000 (00:00 UTC).
			// This is like the way the Unix epoch  works, except the Ripple Epoch is 946684800 seconds after the Unix Epoch.
			// Https://xrpl.org/basic-data-types.html#specifying-time
			CloseTime int64 `json:"close_time"`
			//LedgerData   string   `json:"ledger_data"`
			Transactions []string `json:"transactions"`
		} `json:"ledger"`
		LedgerHash  string `json:"ledger_hash"`
		LedgerIndex int    `json:"ledger_index"`
		Status      string `json:"status"`
		Validated   bool   `json:"validated"`
	} `json:"result"`
}

type transferAmount struct {
	Currency string `json:"currency"`
	Issuer   string `json:"issuer"`
	Value    string `json:"value"`
}

type TxResp struct {
	Result struct {
		Account            string      `json:"Account"`
		Amount             interface{} `json:"Amount"`
		Destination        string      `json:"Destination"`
		DestinationTag     int         `json:"DestinationTag"`
		Fee                string      `json:"Fee"`
		Flags              int64       `json:"Flags"`
		LastLedgerSequence int         `json:"LastLedgerSequence"`
		Sequence           int         `json:"Sequence"`
		SigningPubKey      string      `json:"SigningPubKey"`
		TransactionType    string      `json:"TransactionType"`
		TxnSignature       string      `json:"TxnSignature"`
		Hash               string      `json:"hash"`
		InLedger           int         `json:"inLedger"`
		LedgerIndex        int         `json:"ledger_index"`
		Meta               struct {
			TransactionIndex  int         `json:"TransactionIndex"`
			TransactionResult string      `json:"TransactionResult"`
			DeliveredAmount   interface{} `json:"delivered_amount"`
		} `json:"meta"`
		Status    string `json:"status"`
		Validated bool   `json:"validated"`
	} `json:"result"`
}

type FeeResp struct {
	Result struct {
		Drops struct {
			BaseFee       string `json:"base_fee"`
			MedianFee     string `json:"median_fee"`
			MinimumFee    string `json:"minimum_fee"`
			OpenLedgerFee string `json:"open_ledger_fee"`
		} `json:"drops"`
		Status string `json:"status"`
	} `json:"result"`
}

type accountInfoResp struct {
	Result struct {
		AccountData struct {
			Account           string `json:"Account"`
			Balance           string `json:"Balance"`
			Flags             int    `json:"Flags"`
			LedgerEntryType   string `json:"LedgerEntryType"`
			OwnerCount        int    `json:"OwnerCount"`
			PreviousTxnID     string `json:"PreviousTxnID"`
			PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
			Sequence          int    `json:"Sequence"`
			Index             string `json:"index"`
		} `json:"account_data"`

		Error        string `json:"error"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"result"`
}

type submitParam struct {
	TxBlob string `json:"tx_blob"`
}

type submitResp struct {
	Result struct {
		Error          string `json:"error"`
		ErrorException string `json:"error_exception"`
		ErrorMessage   string `json:"error_message"`

		Accepted            bool                   `json:"accepted"`
		EngineResult        data.TransactionResult `json:"engine_result"`
		EngineResultCode    int                    `json:"engine_result_code"`
		EngineResultMessage string                 `json:"engine_result_message"`
		TxBlob              string                 `json:"tx_blob"`
		Tx                  struct {
			Hash string `json:"hash"`
		} `json:"tx_json"`
	} `json:"result"`
}
