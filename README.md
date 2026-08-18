# crypto-trade-client

Go examples for learning how to query blocks and send native assets on different blockchains. The code is intended for testing and education, not production custody or valuable mainnet funds.

## Build

Build every command under `cmd/` for macOS and the current CPU architecture:

```bash
./build.sh
```

The script writes binaries to `bin/`, including:

- `scanner-core`, `scanner-optimism`, `scanner-polygon`, `scanner-ripple`
- `sender-bitcoin`, `sender-ethereum`, `sender-solana`, `sender-cosmos`

Override the target when cross-compiling:

```bash
GOOS=linux GOARCH=amd64 ./build.sh
```

## Sender configuration

Sender configuration belongs under the ignored `configs/` directory. Never commit a real private key, mnemonic, RPC password, authenticated endpoint, or funded keypair.

Example `configs/chain.yaml` using placeholders:

```yaml
Chains:
  - name: bitcoin
    url: http://127.0.0.1:18443
    rpcUser: <BITCOIN_RPC_USER>
    rpcPassword: <BITCOIN_RPC_PASSWORD>
    wallet: <BITCOIN_CORE_WALLET>
    network: regtest
    fromAddress: <REGTEST_SOURCE_ADDRESS>
    confirmTarget: 6

  - name: ethereum
    url: <ETHEREUM_TESTNET_RPC_URL>
    privateKey: <64_HEX_TEST_PRIVATE_KEY>
    chainID: "<ETHEREUM_TESTNET_CHAIN_ID>"

  - name: solana
    url: <SOLANA_DEVNET_RPC_URL>
    keyPath: configs/solana-id.json

  - name: cosmos
    url: <COSMOS_HUB_TESTNET_GRPC_HOST_PORT>
    mnemonic: <BIP39_TEST_MNEMONIC>
    chainID: <COSMOS_HUB_TESTNET_CHAIN_ID>
    tlsEnabled: true
    gasLimit: 200000
    feeUatom: "5000"
```

Bitcoin `network` accepts `mainnet`, `testnet3`, `regtest`, or `signet`. Cosmos `url` is a gRPC `host:port`, not an HTTP URL.

## Send native assets

Amounts use human-readable decimal units and are converted exactly: BTC supports 8 decimals, ETH 18, SOL 9, and ATOM 6.

```bash
./bin/sender-bitcoin \
  --config configs/chain.yaml \
  --to <BITCOIN_ADDRESS> \
  --amount 0.001 \
  --utxo <TXID>:<VOUT>

./bin/sender-ethereum \
  --config configs/chain.yaml \
  --to <ETHEREUM_ADDRESS> \
  --amount 0.001

./bin/sender-solana \
  --config configs/chain.yaml \
  --to <SOLANA_ADDRESS> \
  --amount 0.01

./bin/sender-cosmos \
  --config configs/chain.yaml \
  --to <COSMOS_ADDRESS> \
  --amount 0.1
```

These commands only validate local input and print a dry-run summary. Append `--broadcast` to explicitly authorize signing and network submission.

For Bitcoin, every `--utxo txid:vout` must belong to `fromAddress`. Bitcoin Core estimates the fee, adds change back to `fromAddress`, signs with the configured wallet, checks mempool acceptance, and then broadcasts. Encrypted wallets must be unlocked separately before running the sender.

## Testing

Some older client tests contact live RPC endpoints or depend on local key files. Inspect tests before running the whole repository. Safe focused checks for the sender implementation are:

```bash
go test ./cmd/sender/...
go test ./clients/bitcoin ./clients/ethereum/client ./clients/solana/gagliardetto ./clients/cosmos
go build ./...
```
