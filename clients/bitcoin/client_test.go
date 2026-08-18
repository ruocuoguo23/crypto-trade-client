package bitcoin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

const testTxID = "0000000000000000000000000000000000000000000000000000000000000001"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSendUsesFixedInputsAndBroadcastsAcceptedTransaction(t *testing.T) {
	fromAddress := regtestAddress(t, 1)
	toAddress := regtestAddress(t, 2)
	var methods []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "rpc-user" || password != "rpc-pass" {
			t.Fatalf("unexpected Basic Auth credentials")
		}
		if r.URL.Path != "/wallet/test-wallet" {
			t.Fatalf("request path = %q", r.URL.Path)
		}

		var request struct {
			ID     uint64            `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, request.Method)

		var result any
		switch request.Method {
		case "getaddressinfo":
			result = map[string]any{"ismine": true, "scriptPubKey": "0014source"}
		case "gettxout":
			result = map[string]any{"scriptPubKey": map[string]any{"hex": "0014source"}}
		case "createrawtransaction":
			var requestedInputs []OutPoint
			if err := json.Unmarshal(request.Params[0], &requestedInputs); err != nil {
				t.Fatalf("decode transaction inputs: %v", err)
			}
			if !reflect.DeepEqual(requestedInputs, []OutPoint{{TxID: testTxID, Vout: 1}}) {
				t.Fatalf("transaction inputs = %#v", requestedInputs)
			}
			var requestedOutputs map[string]json.Number
			if err := json.Unmarshal(request.Params[1], &requestedOutputs); err != nil {
				t.Fatalf("decode transaction outputs: %v", err)
			}
			if requestedOutputs[toAddress] != "0.00100000" {
				t.Fatalf("transaction amount = %q", requestedOutputs[toAddress])
			}
			result = "unsigned-hex"
		case "fundrawtransaction":
			var options map[string]any
			if err := json.Unmarshal(request.Params[1], &options); err != nil {
				t.Fatalf("decode fund options: %v", err)
			}
			if addInputs, ok := options["add_inputs"].(bool); !ok || addInputs {
				t.Fatalf("add_inputs = %#v, want false", options["add_inputs"])
			}
			if options["changeAddress"] != fromAddress {
				t.Fatalf("changeAddress = %#v", options["changeAddress"])
			}
			if options["conf_target"] != float64(3) {
				t.Fatalf("conf_target = %#v", options["conf_target"])
			}
			result = map[string]any{"hex": "funded-hex"}
		case "signrawtransactionwithwallet":
			result = map[string]any{"hex": "signed-hex", "complete": true}
		case "testmempoolaccept":
			result = []map[string]any{{"allowed": true}}
		case "sendrawtransaction":
			result = "broadcast-txid"
		default:
			t.Fatalf("unexpected RPC method %q", request.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"result": result, "error": nil, "id": request.ID})
	})

	client := newTestClient(t, fromAddress, handler)
	txID, err := client.Send(context.Background(), toAddress, 100_000, []OutPoint{{TxID: testTxID, Vout: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if txID != "broadcast-txid" {
		t.Fatalf("txID = %q", txID)
	}
	wantMethods := []string{
		"getaddressinfo", "gettxout", "createrawtransaction", "fundrawtransaction",
		"signrawtransactionwithwallet", "testmempoolaccept", "sendrawtransaction",
	}
	if !reflect.DeepEqual(methods, wantMethods) {
		t.Fatalf("methods = %v, want %v", methods, wantMethods)
	}
}

func TestSendRejectsIncompleteWalletSignature(t *testing.T) {
	fromAddress := regtestAddress(t, 1)
	toAddress := regtestAddress(t, 2)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64 `json:"id"`
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		results := map[string]any{
			"getaddressinfo":               map[string]any{"ismine": true, "scriptPubKey": "0014source"},
			"gettxout":                     map[string]any{"scriptPubKey": map[string]any{"hex": "0014source"}},
			"createrawtransaction":         "unsigned-hex",
			"fundrawtransaction":           map[string]any{"hex": "funded-hex"},
			"signrawtransactionwithwallet": map[string]any{"hex": "partial-hex", "complete": false},
		}
		if request.Method == "testmempoolaccept" || request.Method == "sendrawtransaction" {
			t.Fatalf("%s called after incomplete signature", request.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": results[request.Method], "error": nil, "id": request.ID})
	})

	client := newTestClient(t, fromAddress, handler)
	_, err := client.Send(context.Background(), toAddress, 1, []OutPoint{{TxID: testTxID, Vout: 0}})
	if err == nil || err.Error() != "bitcoin wallet did not fully sign the transaction" {
		t.Fatalf("Send error = %v", err)
	}
}

func TestSendStopsAfterMempoolRejection(t *testing.T) {
	fromAddress := regtestAddress(t, 1)
	toAddress := regtestAddress(t, 2)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64 `json:"id"`
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		results := map[string]any{
			"getaddressinfo":               map[string]any{"ismine": true, "scriptPubKey": "0014source"},
			"gettxout":                     map[string]any{"scriptPubKey": map[string]any{"hex": "0014source"}},
			"createrawtransaction":         "unsigned-hex",
			"fundrawtransaction":           map[string]any{"hex": "funded-hex"},
			"signrawtransactionwithwallet": map[string]any{"hex": "signed-hex", "complete": true},
			"testmempoolaccept":            []map[string]any{{"allowed": false, "reject-reason": "dust"}},
		}
		if request.Method == "sendrawtransaction" {
			t.Fatal("sendrawtransaction called after rejection")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": results[request.Method], "error": nil, "id": request.ID})
	})

	client := newTestClient(t, fromAddress, handler)
	_, err := client.Send(context.Background(), toAddress, 1, []OutPoint{{TxID: testTxID, Vout: 0}})
	if err == nil || err.Error() != "bitcoin mempool rejected transaction: dust" {
		t.Fatalf("Send error = %v", err)
	}
}

func TestParseOutPoint(t *testing.T) {
	got, err := ParseOutPoint(fmt.Sprintf("%s:2", testTxID))
	if err != nil {
		t.Fatal(err)
	}
	if got.TxID != testTxID || got.Vout != 2 {
		t.Fatalf("ParseOutPoint = %#v", got)
	}
	if _, err := ParseOutPoint("not-an-outpoint"); err == nil {
		t.Fatal("ParseOutPoint accepted malformed input")
	}
}

func newTestClient(t *testing.T, fromAddress string, handler http.Handler) *Client {
	t.Helper()
	client, err := NewClient(Config{
		URL: "http://bitcoin.test", RPCUser: "rpc-user", RPCPassword: "rpc-pass", Wallet: "test-wallet",
		Network: "regtest", FromAddress: fromAddress, ConfirmTarget: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Result(), nil
	})}
	return client
}

func regtestAddress(t *testing.T, fill byte) string {
	t.Helper()
	address, err := btcutil.NewAddressWitnessPubKeyHash(bytes.Repeat([]byte{fill}, 20), &chaincfg.RegressionNetParams)
	if err != nil {
		t.Fatal(err)
	}
	return address.EncodeAddress()
}
