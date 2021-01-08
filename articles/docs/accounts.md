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

Note that there's no any "superuser" or any other centralized entity which could move 
tokens from chain accounts without authorisation of its owners.

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

The two most important functions of the `accounts` contract are:

- `deposit`. Allows the caller to deposit its own funds on the chain.
- `withdraw`. Allows the caller to take back its funds from the on-chain account. If the caller
  is a wallet owned by an ordinary address, this means sending the funds from the on-chain
  account back to the address.

By sending requests to the `accounts` contract on a chain, the sender is in
full control on its on-chain funds. Nobody else can move those funds because
the state of the chain can be modified only by the smart contract under the
consensus of the chain's committee of validators.

The `Sandbox` interface provides two methods for a smart contract to interact
with accounts (behind scenes it results in direct calls or on-tangle requests
to the `accounts` smart contract):

- `TransferToAddress` allows the smart contract to transfer its funds to any address on the Tangle
- `TransferCrossChain` allows the smart contract to transfer its funds to any on-chain account on any chain.

## How secure are the on-chain accounts?

On-chain accounts are as secure as the chain they are residing on.

## Node fees

The retainment of node fees uses on-chain accounts following this logic:

- if fees are enabled, they are accrued to the on-chain account of
  `ChainOwnerID`, the _agent ID_ that represents the owner of the chain.
- if fees are disabled, the request tokens (1 mandatory token contained in each request)
  are always accrued to the on-chain account controlled by the sender of the request.
  The requester may withdraw at any time. If never withdrawn and deposited separately,
  the account will contain a number of iotas equal to the numebr of requests sent by
  that requester.
