# The `accounts` contract

The `accounts` contract is one of the [core contracts](overview.md) on each ISCP
chain.

The function of the `accounts` contract is to keep in its state a consistent ledger of
on-chain accounts for the agents that control them. There are two types of agents who can do it:
L1 addresses and smart contracts.

The `accounts` contract provides functions to deposit and withdraw tokens.
It also provides information about the assets deposited on the chain.  
Note that the ledger of accounts on the chain is consistently maintained behind scenes by the VM.
The `accounts` contract provides a front-end of authorized access to those
accounts for outside usersof the chain.

### Entry Points

* **deposit** moves tokens attached as a transfer to a target account on the
  chain. If the agent ID parameter `a` is specified the target account is the
  one controlled by that agent ID. Otherwise, the target account is the one
  controlled by the caller (this makes sense only if it is a request, not if it
  is an on-chain call).

* **withdraw** moves all tokens from the caller's on-chain account to another
  chain, or to an address on L1. It cannot be used to move tokens within the
  current chain.

* **harvest** moves tokens from the common (default( account controlled by the chain owner to the proper owner's
  account on the same chain. Only authorised to whoever is an owner of the chain.

### Views

* **accounts** returns a list of all non-empty accounts in the chain as a list
  of serialized `agent IDs`.

* **balance** returns the colored token balances that are controlled by the
  `agent ID` that was specified in the call parameters. It returns the
  balances as a dictionary of `color: amount` pairs.

* **totalAssets** - Returns the colored balances controlled by the chain.
  They are always equal to the sum of all on-chain accounts, color-by-color.

