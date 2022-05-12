---
description: An overview of the IOTA Smart Contracts architecture.
image: /img/multichain.png
keywords:
- smart contracts
- architecture
- Layer 2
- Layer 1
- explanation
---

# ISC Architecture

With IOTA Smart Contracts, anyone can start their own chain and invite others to validate it. Together, these chains are called Layer 2.

Each chain has its own state and smart contracts that cause this state to change. As validators execute smart contracts, they tally these changes and write them into the chain. Each time they update the state, they collectively agree on a new state and commit to it by publishing its hash to Layer 1, the Tangle.

Each Layer 2 chain is functionally equivalent to, say, Ethereum, but all chains can communicate with Layer 1 and each other, making ISC a more intricate protocol.

![A diagram with multiple smart contract chains. Each is functionally equivalent to the Ethereum blockchain, but they also communicate to Layer 1 and each other.](/img/multichain.png "Click to see the full-size image.")

*IOTA Smart Contacts multichain architecture.*

The comprehensive overview of architectural design decisions of IOTA Smart Contracts can be found in the 
[ISC white paper](https://files.iota.org/papers/ISC_WP_Nov_10_2021.pdf).