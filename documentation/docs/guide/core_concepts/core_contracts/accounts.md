# The `accounts` Contract

The `accounts` contract is one of the [core contracts](overview.md) on each ISCP
chain.

The `accounts` contract keeps a consistent ledger of on-chain accounts in its state for the agents that control them. There are two types of agents who can do it: L1 addresses and smart contracts.

## Entry Points

The `accounts` contract provides functions to deposit and withdraw tokens, and also provides information about the assets deposited on the chain.  

Note that the ledger of accounts on the chain is consistently maintained behind scenes by the VM.

### deposit

Moves tokens attached as a transfer to a target account on the chain. By default, the funds are deposited to the caller account. Optionally, a different target account can be specified with the agent ID parameter `a`.

### withdraw

Moves all tokens from the caller's on-chain account to another chain, or to an address on L1. It cannot be used to move tokens within the current chain.

### harvest

Moves tokens from the common "default" account controlled by the chain owner, to the proper owner's account on the same chain. This entry point is only authorised to whoever owns the chain.

## Views

The `accounts` contract provides a front-end of authorized access to those accounts for users outside the chain.

### accounts

Returns a list of all non-empty accounts in the chain as a list of serialized `agent IDs`.

### balance

Returns the colored token balances that are controlled by the `agent ID` that was specified in the call parameters. It returns the balances as a dictionary of `color: amount` pairs.

### totalAssets

Returns the colored balances controlled by the chain. They are always equal to the sum of all on-chain accounts, color-by-color.
