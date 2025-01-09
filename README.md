# Forta JSON-RPC Proxy

This repository contains a JSON-RPC proxy service intended for running in front of a JSON-RPC endpoint in order to provide a Forta Firewall flavor.

This service is currently under development and is subject to change. See the bottom section for manual testing.

## Functionality

The HTTP proxy handler `ServeHTTP()` defined in `service/proxy.go` treats received JSON-RPC requests in three different categories:

### Wrapped methods

- **eth_call:** Forta Firewall-protected contracts execute checkpoints that revert a transaction by default, unless the transaction is supported with an off-chain attestation written on-chain. This causes a revert-related friction in user wallets. To make this revert go away, `eth_call` is wrapped to add a special state override argument supported by the Forta Firewall contracts on the target chain.

- **eth_estimateGas:** Same as above.

- **eth_sendRawTransaction:** User transaction is frontran in this method. First, the user transaction is checked against a Forta Attester. If the Forta Attester gives back an attestation transaction, then one of the two flows take place:
	- _Ethereum mainnet:_ A transaction bundle is sent to a block builder API (`eth_sendBundle`).
	- _Other chains:_ Attestation transaction is sent to the proxy target, receipt is awaited, and then the user transaction is sent to the proxy target.

These methods are wrapped in `service/service.go` and registered to `eth` namespace to the JSON-RPC server in `service/proxy.go`.

### Proxied methods

These methods are other bunch of methods which are fundamental to the functionality of a wallet app but are not wrapped. Whenever intercepted, these methods are proxied directly to the target endpoint.

- net_version
- eth_chainId
- eth_getBalance
- eth_getTransactionCount
- eth_getBlockByNumber
- eth_getBlockByHash
- eth_blockNumber
- eth_getCode
- eth_gasPrice
- eth_getTransactionReceipt

### Authorized methods

Any method outside of the wrapped and proxied methods are restricted to power users of the API with the help of an API key, because of the potentially heavy cost of these methods. The API key mechanism can be improved later to support multiple API keys flexibly.

## Testing

Normally, the proxy server should be started through `main.go` but there is an alternative build for testing, in `testing/testproxy/main.go`. It is almost the same, except, the attester is included as a fake one in the same build, instead of a remote one. This attester works with a fake security validator and protects a dummy contract which can be found in `testing/contracts`. The high level steps are:
- Deploy contracts with a simple tool like Remix IDE (or if you want scripting, you could do Foundry or Hardhat).
- Note down the addresses of the contracts.
- Create a `.env` file and provide all required env vars, defined in `testing/testproxy/main.go`.
- Run `testing/testproxy/main.go`.
