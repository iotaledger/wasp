# On-chain accounts

ISCP introduces the concept of _on-chain account_. Each chain maintains a list
of pairs: `<agent ID>: <colored balance>`.  Each pair is an account with its
colored balances.

**Any agent ID on the ISCP network may have an account on any chain**.  In
other words, any smart contract and any ordinary address on the network can
have account on any chain.

ISCP ensures that the tokens owned by the chain address may be moved to another
location only by the entity represented by the corresponding agent ID.  The
system requires cryptographically secure authorization to move funds between
on-chain accounts. 

Note that there isn't any "superuser" or any other centralized entity which could move 
tokens from chain accounts without authorization of its owners.

Some corollaries:

- Any entity may move its tokens seamlessly from an address on the tangle to the account
  controlled by the same address on any chain.
- Anyone can send tokens to the account of any smart contract on any chain.
- An address may, at any time, withdraw it tokens from the chain, transferring
  them to the address on the Tangle.
- A contract may keep its funds on its native chain or on any other chain.

## How on-chain accounts work

Each chain contains several built-in smart contracts, called _core contracts_.
One of those _core_ contracts is the `accounts` contract, which handles the whole
account machinery for each chain.

All funds belonging to the contracts on some chain are actually owned by the
chain address in the IOTA ledger (level 1).  The `accounts` contract keeps the
account balance for each agent ID in the chain state (level 2).

Funds are moved to and from any on-chain accounts by calling `accounts`
functions on that chain.  The `accounts` contract does all the ledger
accounting and guarantees security.

In each call, the `accounts` contract securely knows the agent ID of the
_caller_ (either an ordinary wallet or another smart contract) and authorizes
the call.  For example, a call to the `withdraw` function will only be
authorized if called from the _agent ID_ of the owner of the account.

The most important functions of the `accounts` contract are:

- `deposit`. Allows the caller to deposit its own funds to any target account on the chain.
- `withdrawToAddress`. Allows a L1 address (a wallet) to take funds from its on-chain account back to the address. 
- `withdrawToChain`. Allows a smart contract to take back its funds from another chain to its native chain. 

By sending requests to the `accounts` contract on a chain, the sender is in
full control of its on-chain funds. 

The `Sandbox` interface provides `TransferToAddress` method for the smart contract 
to transfer its funds to any address on the Tangle.

For more information see [accounts contract](../guide/core_concepts/core_contracts/accounts.md).

## How secure are the on-chain accounts?

On-chain accounts are as secure as the chain they reside on.

## Node fees

The node fees are charged by using the following logic:

- if fees are enabled, they are accrued to the on-chain account of
  `ChainOwnerID`, the _agent ID_ that represents the owner of the chain.
- if fees are disabled, the request tokens (1 mandatory token contained in each request)
  are always accrued to the on-chain account controlled by the sender of the request.
  The requester may withdraw it at any time. 
