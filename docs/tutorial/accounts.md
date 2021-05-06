# The `accounts` contract

The `accounts` contract is one of the [core contracts](coresc.md) on each ISCP
chain.

The function of the `accounts` contract is to keep a consistent ledger of
on-chain accounts for the agents that control them: L1 addresses and smart
contracts.

The `accounts` contract provides functions to deposit and withdraw tokens, and
also provides information about the assets deposited on the chain. Note that its
ledger of accounts is consistently maintained behind scenes by the VM.
The `accounts` contract provides a front-end of authorized access to those
accounts for outside users.

### Entry Points

* **deposit** - Moves tokens attached as a transfer to a target account on the
  chain. If the agent ID parameter `a` is specified the target account is the
  one controlled by that agent ID. Otherwise, the target account is the one
  controlled by the caller (this makes sense only if it is a request, not if it
  is an on-chain call).

* **withdraw** - Moves all tokens from the caller's on-chain account to another
  chain, or to an address on L1. It cannot be used to move tokens within the
  current chain.

### Views

* **accounts** - Returns a list of all non-empty accounts in the chain as a list
  of `agent IDs`.

* **balance** - Returns the colored token balances that are controlled by the
  agent ID `a` that was specified in the call parameters. It returns the
  balances as a dictionary of `color: amount` pairs.

* **totalAssets** - Returns the colored token balances controlled by the chain.
  They are always equal to the sum of all on-chain accounts.

