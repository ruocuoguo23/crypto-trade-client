# AGENTS.md

## Repository purpose

This repository contains Go examples for learning and testing blockchain client development. Its main use cases are:

- connecting to RPC nodes on different blockchains;
- querying blocks, transactions, balances, and chain state;
- constructing, signing, and broadcasting transactions;
- polling or scanning newly produced blocks;
- comparing SDKs and implementation patterns across chains.

The code is educational and experimental. Do not assume it is production-ready, audited, highly available, or safe for managing valuable assets.

## Technology and layout

- Go version: 1.21.
- `clients/<chain>/`: chain-specific clients and examples.
- `clients/ethereum/client/`: shared EVM client and signing logic.
- `cmd/scanner/<chain>/`: executable block-scanning examples.
- `cmd/sender/`: transaction-sending command entry point.
- `common/`: shared configuration, cryptography, RPC, HTTP, and utility packages.

Keep chain-specific behavior in its chain package. Share code only when the chains genuinely use the same protocol or implementation, as with the EVM clients.

## Change guidelines

- Prefer small, readable examples over framework-like abstractions.
- Preserve the repository's learning value: make transaction construction, signing, RPC calls, and block processing easy to follow.
- Do not create a universal cross-chain interface unless the task explicitly requires one.
- Follow existing Go naming and package conventions, and run `gofmt` on changed Go files.
- Wrap returned errors with useful operation context, while preserving the original error with `%w` when appropriate.
- Use `context.Context` for network operations and add reasonable timeouts to new commands or long-running requests.
- Document chain-specific units and assumptions, such as wei, satoshi, lamports, block finality, address format, and required confirmations.
- Avoid unrelated refactors or formatting changes.

When adding support for a chain or a new operation, prefer this minimal shape:

1. Add the RPC/client behavior under `clients/<chain>/`.
2. Add focused tests for deterministic logic.
3. Add a small command example under `cmd/` only when an executable demonstration is useful.
4. Explain required RPC configuration and whether the example uses localnet, testnet/devnet, or mainnet.

## Blockchain safety

- Never commit private keys, seed phrases, mnemonics, keystore files, API tokens, authenticated RPC URLs, or funded test accounts.
- Never print secrets or full signed transaction payloads when they may expose sensitive material.
- Read credentials from environment variables or files under the ignored `configs/` directory. Example values must be obviously fake.
- Default new examples to a local network or public testnet/devnet. Mainnet use must be explicit.
- Treat transaction broadcasting as an external write. Before running a command or test that sends a transaction, verify the network, sender, recipient, asset, amount, fee settings, and expected side effects.
- Do not execute tests that require a private key, consume funds, or broadcast transactions unless the user explicitly asks for that action.
- Use integer types such as `big.Int` for on-chain amounts. Do not use floating-point values for asset or fee calculations.
- Validate chain IDs, addresses, public keys, block heights, and transaction identifiers before using them.
- For scanners, account for duplicate blocks, skipped heights, RPC retries, reorganizations, and the chain's finality model when the example's scope requires it.

## Testing and verification

Some existing client tests contact live RPC endpoints, depend on local key files, or may broadcast transactions. Inspect a test before running it; do not blindly run the entire suite.

Use the narrowest safe verification that covers the change:

```bash
gofmt -w <changed-go-files>
go test ./common/...
go test ./path/to/changed/package
go build ./...
```

For new network-dependent tests:

- separate deterministic unit tests from integration tests;
- obtain endpoints and credentials from environment variables or ignored local configuration;
- skip with a clear message when required integration settings are absent;
- never send a transaction as an incidental part of a read-only test;
- state the required network and any possible cost in the test name or nearby documentation.

## Completion criteria

A change is complete when:

- it is scoped to the requested chain or shared component;
- changed Go code is formatted and builds;
- relevant safe tests pass;
- network assumptions and side effects are clear;
- no secrets or machine-specific paths were added;
- any unrun integration test is called out with the reason it was not run.
