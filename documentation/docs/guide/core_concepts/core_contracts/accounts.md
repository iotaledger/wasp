---
description: The `accounts` contract keeps the ledger of on-chain accounts.
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

# The `accounts` Contract

The `accounts` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts
chain.

This contract keeps a consistent ledger of on-chain accounts in its state, i.e. [the L2 ledger](../accounts/how-accounts-work).

---

## Entry Points

The `accounts` contract provides functions to deposit and withdraw tokens, information about the assets deposited on the chain, as well as functionality to create/utilize foundries.  

### `deposit()`

A no-op that has the side effect of crediting any transferred tokens to the sender's account.

:::note
As with every call, the gas fee is debited from the L2 account right after
executing the request.
:::

### `withdraw()`

Moves tokens from the caller's on-chain account to the caller's L1 address.
The amount of tokens to be withdrawn must be specified via the allowance of the request.

:::note
A call to withdraw means that a L1 output will be created.
Because of this, the withdrawn amount must be able to cover the L1 storage deposit, otherwise it will fail.
:::

### `transferAllowanceTo(a AgentID, c ForceOpenAccount)`

Moves the specified allowance from the sender's L2 account to the given L2 account on the chain.

Parameters:

- `a` (`AgentID`): The target L2 account
- `c` (optional `bool` - default: `false`): If the target Agent ID doesn't yet have funds on the chain, `c = true` specifies that the account should be created; otherwise the call fails.

### `harvest(f ForceMinimumBaseTokens)`

Only the chain owner can call this entry point, which moves all tokens from the chain [common account](../accounts/the-common-account) to the sender's L2 account.

Parameters:

- `f` (optional `uint64` - default: `MinimumBaseTokensOnCommonAccount`): specifies the amount of base tokens to leave in the common account.

### `foundryCreateNew(t TokenScheme) s SerialNumber`

Creates a new foundry with the specified token scheme, and assigns the foundry to the sender.

Parameters:

- `t` ([`iotago::TokenScheme`](https://github.com/iotaledger/iota.go/blob/develop/token_scheme.go)): The token scheme for the new foundry.

The storage deposit for the new foundry must be provided via allowance (only the minimum required will be used).

Returns:

- `s` (`uint32`): The serial number of the newly created foundry

### `foundryModifySupply(s SerialNumber, d SupplyDeltaAbs, y DestroyTokens)`

Mints or destroys tokens for the given foundry, which must be controlled by the caller.

Parameters:

- `s` (`uint32`): The serial number of the foundry
- `d` (positive `big.Int`): Amount to mint or destroy
- `y` (optional `bool` - default: `false`) Whether to destroy tokens (`true`) or not (`false`)

When minting new tokens, the storage deposit for the new output must be provided via allowance.

When destroying tokens, the tokens to be destroyed must be provided via allowance.

### `foundryDestroy(s SerialNumber)`

Destroys a given foundry output on L1, reimbursing the storage deposit to the caller.
The foundry must be owned by the caller.

:::warning
This operation cannot be reverted.
:::

Parameters:

- `s` (`uint32`): The serial number of the foundry

---

## Views

### `balance(a AgentID)`

Returns the fungible tokens owned by the given Agent ID on the chain.

Parameters:

- `a` (`AgentID`): The account Agent ID

Returns:

A map of [`TokenID`](#tokenid) => `big.Int`. The L1 base tokens is represented by an empty Token ID (a key with length 0).

### `balanceBaseToken(a AgentID)`

Returns amount of base tokens owned by any AgentID `a` on the chain.

Parameters:

- `a` (`AgentID`): The account Agent ID

Returns:

- `B` (`uint64`): The amount of base tokens in the account

### `balanceNativeToken(a AgentID, N TokenID)`

Returns the amount of native tokens with Token ID `N` owned by any AgentID `a`  on the chain.

Parameters:

- `a` (`AgentID`): The account Agent ID
- `N` ([`TokenID`](#tokenid)): The Token ID

Returns:

- `B` (`big.Int`): The amount of native tokens in the account

### `totalAssets()`

Returns the sum of all fungible tokens controlled by the chain.

Returns:

A map of [`TokenID`](#tokenid) => `big.Int`. The L1 base tokens is represented by an empty Token ID (a key with length 0).

### `accounts()`

Returns a list of all agent IDs that own assets on the chain.

Returns: a map of `AgentID` => `0xff`.

### `getNativeTokenIDRegistry()`

Returns a list of all native tokenIDs that are owned by the chain.

Returns: a map of [`TokenID`](#tokenid) => `0xff`

### `foundryOutput(s FoundrySerialNumber)`

Returns the output corresponding to the foundry with Serial Number `s`.

Returns:

- `b`: [`iotago::FoundryOutput`](https://github.com/iotaledger/iota.go/blob/develop/output_foundry.go)

### `accountNFTs(a AgentID)`

Returns the NFT IDs for all NFTs owned by the given account.

Parameters:

- `a` (`AgentID`): The account Agent ID

Returns:

- `i` ([`Array16`](https://github.com/dessaya/wasp/blob/develop/packages/kv/collections/array16.go) of [`iotago::NFTID`](https://github.com/iotaledger/iota.go/blob/develop/output_nft.go)): The NFT IDs owned by the account

### `nftData(z NFTID)`

Returns the data for a given NFT with ID `z` that is on the chain.

Returns:

- `e`: [`NFTData`](#nftdata)

### `getAccountNonce(a AgentID)`

Returns the current account nonce for a give AgentID `a`.
The account nonce is used to issue off-ledger requests.

Parameters:

- `a` (`AgentID`): The account Agent ID

Returns:

- `n` (`uint64`): The account nonce

## Schemas

### `TokenID`

```
TokenID = [38]byte
```

### `NFTData`

`NFTData` is encoded as the concatenation of:

- The issuer ([`iotago::Address`](https://github.com/iotaledger/iota.go/blob/develop/address.go))
- The NFT metadata: the length (`uint16`) followed by the data bytes
- The NFT owner (`AgentID`)

