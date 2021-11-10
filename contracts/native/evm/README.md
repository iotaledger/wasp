# EVM support

This package and subpackages contain the code for the `evmchain` and `evmlight`
smart contracts, which allow to execute Ethereum VM (EVM) code on top of the
ISCP chain, thus adding support for Ethereum smart contracts.

## EVM flavors

There are two EVM 'flavors', each one being a native ISCP smart contract. You
must decide which one to use when deploying a new EVM chain:

### `evmchain`

The `evmchain` implementation emulates a full Ethereum blockchain, as if it
were an Ethereum 'node', storing the full history of blocks and past states.

Pros:
- More likely to be compatible with existing EVM tools (e.g.: Metamask, Remix,
  Hardhat, etc.) that expect a 'real' Ethereum blockchain.

Cons:
- Inefficient: spends time and space calculating the Merkle tree, verifying
  blocks, storing past states, etc, when none of this is necessary in an ISCP
  contract.

### `evmlight`

The `evmlight` implementation is a more efficient solution for EVM support. It
stores only the current EVM state in raw form, and after running an Ethereum
transaction, it:

- Updates the EVM state
- Stores the transaction and receipt for future reference. Only the latest N
  transactions/receipts are stored, to avoid using unlimited space.

Pros:
- More space/time efficient than `evmchain`
- Potentially easier to integrate with ISCP (still in experimental phase)

Cons:
- Less support for existing EVM tools that expect a 'real' Ethereum blockchain.
  There is still partial support for some EVM tools. YMMV.

## Enabling / disabling EVM

EVM support is provided by the `evmchain` and `evmlight` native contracts, and
as such it needs to be enabled at compile time. **EVM support is enabled by
default, so no special action is needed.**

EVM support inflates the `wasp` and `wasp-cli` binaries by several MB. If this
is a problem and you don't need EVM support, you can disable it at compile
time by providing the `-tags noevm` flag to the Go compiler. For example:

```
go install -tags noevm ./...
```

## Deploy

You can use `wasp-cli` to deploy the `evmchain` or `evmlight` contract (given that you
already have access to an ISCP chain and have deposited some funds into your
on-chain account):

```
wasp-cli chain evm deploy --alloc 0x71562b71999873DB5b286dF957af199Ec94617F7:115792089237316195423570985008687907853269984665640564039457584007913129639927
```

The `--alloc` parameter specifies the genesis allocation for the EVM chain,
with syntax `address:wei [address:wei ...]`.

By default the `evmchain` contract will be deployed; you can change this with
`--evm-flavor evmlight`.

## JSON-RPC

Once your EVM chain is deployed, you can use the `wasp-cli chain evm jsonrpc`
command to start a JSON-RPC server. This will allow you to connect any standard
Ethereum tool, like Metamask.

Note: If you are using `evmlight` you should run the JSON-RPC server with
`--name evmlight`.

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

## Predictable block time

Some EVM contracts depend on blocks being minted periodically with regular
intervals. ISCP does not support that natively, so by default a new EVM block
is minted every time an ISCP batch is executed that contains at least one EVM
transaction. In other words, by default no EVM blocks will be minted until an
EVM transaction is received.

However, the `evmlight` implementation supports emulating predictable block
times. To enable this feature, just pass the `--block-time n` flag when
deploying the EVM chain with `wasp-cli chain evm deploy`, where `n` is
the desired average amount of seconds between blocks.

Note that this may change the behavior of JSON-RPC functions that query the
EVM state (e.g. `getBalance`), since `evmlight` is not able to store the state
in both the latest minted block and the pending block. These functions will
always return the state computed after accepting the latest transaction (i.e.
the state of the pending block).
