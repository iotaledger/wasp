---
description: 'The evm core contract provides the necessary infrastructure to accept Ethereum transactions and execute
EVM code.'
image: /img/logo/WASP_logo_dark.png
keywords:

- smart contracts
- core
- root
- initialization
- entry points
- fees
- ownership
- views
- reference

---

# The `evm` Contract

The `evm` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts chain.

The `evm` core contract provides the necessary infrastructure to accept Ethereum transactions and execute EVM code.
It also includes the implementation of the [ISC Magic contract](../../evm/magic.md).

:::note
For more information about how ISC supports EVM contracts, refer to the [EVM](../../evm/introduction.md) section.
:::

---

## Entry Points

Most entry points of the `evm` core contract are meant to be accessed through the JSON-RPC service provided
automatically by the Wasp node so that the end users can use standard EVM tools like [MetaMask](https://metamask.io/).
We only list the entry points not exposed through the JSON-RPC interface in this document.

### `init()`

Called automatically when the ISC is deployed.

Some parameters of the `evm` contract can be specified by passing them to the
[`root` contract `init` entry point](root.md#init):

- `evmg` (optional [`GenesisAlloc`](#genesisalloc)): The genesis allocation. The balance of all accounts must be 0.
- `evmgl` (optional `uint64` - default: 15000000): The EVM block gas limit (EVM gas units)
- `evmbk` (optional `int32` - default: keep all): Amount of EVM blocks to keep in the state.
- `evmchid` (optional `uint16` - default: 1074): EVM chain iD

  :::caution

  Re-using an existing Chain ID is not recommended and can be a security risk. For serious usage, register a unique
  Chain ID on [Chainlist](https://chainlist.org/) and use that instead of the default. **It is not possible to change
  the EVM chain ID after deployment.**

  :::

- `evmw` (optional [`GasRatio`](#gasratio) - default: `1:1`): The ISC to EVM gas ratio.

### `setGasRatio`

Changes the ISC : EVM gas ratio.

#### Parameters

- `w` ([`GasRatio`](#gasratio)): The ISC : EVM gas ratio.

---

## Views

### `getGasRatio`

Returns the ISC : EVM gas ratio.

#### Returns

- `r` ([`GasRatio`](#gasratio)): The ISC : EVM gas ratio.

---

## Schemas

### `GenesisAlloc`

`GenesisAlloc` is encoded as the concatenation of:

- Amount of accounts `n` (`uint32`).
- `n` times:
    - Ethereum address (`[]byte` prefixed with `uint32` size).
    - Account code (`[]byte` prefixed with `uint32` size).
    - Amount of storage key/value pairs `m`(`uint32`).
    - `m` times:
        - Key (`[]byte` prefixed with `uint32` size).
        - Value(`[]byte` prefixed with `uint32` size).
    - Account balance (must be 0)(`[]byte` prefixed with `uint32` size).
    - Account nonce  (`uint64`).
    - Account private key (may be used for tests)(`uint64`).

### `GasRatio`

ISC : EVM gas ratio is expressed as an `a : b` ratio, where `<ISC gas> = <EVM gas> * <a> / <b>`.

`GasRatio` is encoded as the concatenation of the two `uint32` values `a` & `b`.



