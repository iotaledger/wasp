---
keywords:
- Validators
- consensus
- state update
description: Each chain is run by a network of validator nodes which run a consensus on the chain state update.
image: /img/logo/WASP_logo_dark.png
---
# Validators

Each chain is run by a network of validator nodes, which run a consensus on the chain state updates. The [Wasp](https://github.com/iotaledger/wasp) node is an implementation of the validator node. The validators of the chain form a committee, a bound together closed set of nodes. The committee of the chain may change, allowing new validators and validator nodes to be added or replaced. This also makes the chain itself agnostic to its validators (the committee).

Only when a supermajority of the validators (the quorum) of a chain reaches [consensus](./consensus.md), a new state update can be signed, which unlocks the AliasOutput for the chain and produces the next state UTXO.
