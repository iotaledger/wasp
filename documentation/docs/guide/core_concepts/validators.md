---
description: Each chain is run by a network of validator nodes which run a consensus on the chain state update.
image: /img/logo/WASP_logo_dark.png
keywords:
- validators
- validator nodes
- access nodes
- consensus
- state update
- explanation
---
# Validators

Each chain is run by that chain's *committee of validators*. This committee owns a key that is split between all of its validators. Each key share is useless on its own, but a collective signature gives validators full control over the chain.

The committee of validators is responsible for executing the smart contracts in the chain, and thus calculating a _state update_.
All validators execute exactly the same code and reach consensus on the state update.
Once the next state is computed and validated, it is committed on each validator's database, a new _block_ is added to the chain (containing the state mutations), and the _state hash_ is saved in the L1 ledger.

The committee of validators can be rotated depending on the governance model.
By rotating the committee of validators, validators can be deleted, added, or replaced.

ISC does not define how to select validators to form a committee: it could be a solitary choice of the chain's owner, or it could be a [public competition](https://wiki.assembly.sc/learn/introduction/) between candidates.
ISC does not define how validators are rewarded, either.

## Access nodes

It is possible to have some nodes that act as _access nodes_ to the chain, without being be part of the committee of validators.
All nodes in the subnet (validators and non-validators) are connected to each other through statically assigned trust relationships, and each node is also connected to the IOTA L1 node in order to receive updates on the chain’s L1 account.

Any node can optionally provide access to smart contracts for external callers, allowing them to:
* Query the state of the chain (i.e. _view calls_)
* Send off-ledger requests directly to the node (instead of sending an on-ledger request as a L1 transaction)

It is common for validator nodes to be part of a private subnet, and have only a group of access nodes exposed to the outside world, protecting the committee from external attacks.

The management of validator and access nodes is done through the [`governance` core contract](./core_contracts/governance).
