# The `evmchain` smart contract

The `evmchain` smart contract emulates an Ethereum blockchain on top of the
ISCP chain, allowing to run Ethereum smart contracts.

## EVM support

The `evmchain` contract is implemented as a native contract, and as such it
needs to be enabled at compile time. **EVM support is enabled by default, so
no special action is needed.**

EVM support inflates the `wasp` and `wasp-cli` binaries by several MB. If this
is a problem and you don't need EVM support, you can disable it at compile
time by providing the `-tags noevm` flag to the Go compiler. For example:

```
go install -tags noevm ./...
```

## Deploy

You can use `wasp-cli` to deploy the `evmchain` contract (given that you
already have access to an ISCP chain and have deposited some funds into your
on-chain account):

```
wasp-cli chain evm deploy --alloc 0x71562b71999873DB5b286dF957af199Ec94617F7:115792089237316195423570985008687907853269984665640564039457584007913129639927
```

The `--alloc` parameter specifies the genesis allocation for the EVM chain,
with syntax `address:wei [address:wei ...]`.

## JSON-RPC

Once your EVM chain is deployed, you can use the `wasp-cli chain evm jsonrpc`
command to start a JSON-RPC server. This will allow you to connect any standard
Ethereum tool, like Metamask.

## Complete example using `wasp-cluster`

In terminal #1, start a cluster:

```
wasp-cluster start -d
```

In terminal #2:

```
# initialize a private key and request some funds
wasp-cli init
wasp-cli request-funds

# deploy an ISCP chain, deposit some funds to be used for fees
wasp-cli chain deploy --chain=mychain --committee=0,1,2,3 --quorum 3
wasp-cli chain deposit IOTA:1000

# deploy an EVM chain
wasp-cli chain evm deploy --alloc 0x71562b71999873DB5b286dF957af199Ec94617F7:115792089237316195423570985008687907853269984665640564039457584007913129639927
```

Finally we start the JSON-RPC server:

```
wasp-cli chain evm jsonrpc
```
