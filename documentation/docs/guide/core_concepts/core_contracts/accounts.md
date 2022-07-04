---
description: The `accounts` contract keeps a consistent ledger of on-chain accounts in its state for the agents that control them. There are two types of agents who can do it, L1 addresses and smart contracts.
image: /img/logo/WASP_logo_dark.png
keywords:
- core contracts
- accounts
- deposit
- withdraw
- assets
- balance
- reference
--- 

// TODO  update <https://stardust.iota-community.org> links to the wiki

# The `accounts` Contract

The `accounts` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts
chain.

This contract keeps a consistent ledger of on-chain accounts in its state, establishing what is owned by who.
There are three types of agents who can own assets on the chain: L1 addresses, ISC smart contracts and EVM smart contracts.
Any agent can own L1 assets: tokens, NFTs and foundries

---

## Entry Points

The `accounts` contract provides functions to deposit and withdraw tokens, information about the assets deposited on the chain, as well as functionality to create/utilize foundries.  

### - `deposit()`

Credits any transfered tokens to the sender's account.

### - `withdraw()`

Moves tokens from the caller's on-chain account to any external L1 address (can be an Agent on another chain).
The amount of tokens to be withdrawn must be specified via allowance in the request.

:::note
A call to withdraw means that a L1 output will be created, because of this, the withdrawn amount must be able to cover the L1 [storage deposit](https://stardust.iota-community.org/introduction/develop/introduction/what_is_stardust#storage-deposit-system), otherwise it will fail.
:::

### - `transferAllowanceTo(a AgentID)`

Credits the specified allowance to any AgentID (`a`) on the chain.

:::note
If the target AgentID doesn't yet have funds on the chain, an optional boolean parameter (`c`) must specified to signal for an account to be created.
:::

### - `harvest()`

Moves tokens from the common account controlled by the chain owner, to the proper owner's account on the same chain. This entry point is only authorised to whoever owns the chain.

### - `foundryCreateNew(t TokenScheme) s SerialNumber`

Creates a new foundry with the specified [token scheme](https://stardust.iota-community.org/introduction/develop/protocol/foundry) `t`. The new foundry is created under the controller of the request sender.
The serial number `s` of the newly created foundry will be returned.

:::note
The [storage deposit](https://stardust.iota-community.org/introduction/develop/introduction/what_is_stardust#storage-deposit-system) for the new foundry must be provided via allowance (only the minimum required will be used).
:::

### - `foundryModifySupply(s SerialNumber, d SupplyDeltaAbs, y DestroyTokens)`

Inflates (mints) or shrinks supply of token by the foundry, controlled by the caller.
The following parameters must be provided:

- the target foundry serial number `s`
- SupplyDeltaAbs `d` specifies by which amount the supply should increase or decrease (specified as a big.int), this is an absolute value
- DestroyTokens `y` is a boolean that specifies whether to destroy tokens or not (defaults to `false`)

When minting new tokens, the storage deposit for the new output must be provided via allowance.

When destroying tokens, the tokens to be destroyed must be provided via allowance.


### - `foundryDestroy(s SerialNumber)`

Destroys a given foundry output on L1, reiburses the [storage deposit](https://stardust.iota-community.org/introduction/develop/introduction/what_is_stardust#storage-deposit-system) to the caller. (Can only succeed if the foundry is owned by the caller)

:::warning
This operation cannot be reverted
:::

---

## Views

The `accounts` contract provides ways to query information about chain accounts.

### - `balance(a AgentID)`

Returns the fungible tokens owned by any AgentID `a`.


### - `accountNFTs()`

Returns the NFTs owned by any AgentID `a`.

### - `balanceBaseToken()`



### - `balanceNativeToken()`

### - `totalAssets()`

Returns the colored balances controlled by the chain. They are always equal to the sum of all on-chain accounts, color-by-color.

### - `accounts()`

Returns a list of all non-empty accounts in the chain as a list of serialized `agent IDs`.

### - `getAccountNonce()`

### - `getNativeTokenIDRegistry()`

### - `foundryOutput()`



### - `nftData()`
