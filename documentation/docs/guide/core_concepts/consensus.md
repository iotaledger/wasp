---
keywords:
- ISCP
- Smart Contracts
- Consensus
description: ISCP Consensus
image: /img/logo/WASP_logo_dark.png
---
# Consensus

## Consensus on computations

Even if all committee nodes are honest (i.e. they have no malicious intent), there are factors which may make each node see things differently. This can lead to different inputs to the same program on different nodes and, consequently, to different results.

There are several possible reasons for such an apparently non-deterministic outcome.

Each committee node has its own access to the UTXO ledger, i.e. committee nodes are usually connected to different IOTA nodes. The reason for this is to not make access to the UTXO ledger a single point of failure, i.e. we also want access to the Tangle to be distributed. This may often lead to a slightly different perception of some aspects of the ledger, for example of the token balance in a particular address. Also, each node has its own local clock and those clocks may be slightly skewed, so there isn’t an objective time for nodes.

The requests (UTXOs) may reach Wasp nodes in an arbitrary order and with arbitrary delays (even if these are usually close to the network latency).

Before starting calculations, nodes are required to have consensus on the following:

- The current state of the chain i.e. on the state output
- Timestamp to be used for the next state transaction
- Ordered batch of requests to be processed
- Address where node fees for processing the request must be sent (if enabled)
- Mana pledge targets

In order to achieve a bigger throughput, the committee picks requests from the on-ledger backlog and processes requests in batches, not one by one. This means the committee has to have a consensus on the batch of the requests and the order of the requests in the batch. After at least a quorum of committee nodes have a consensus on the above, honest committee members will always produce identical results of calculations.

## Proof of consensus

Suppose a quorum of committee nodes has reached consensus on inputs and produced identical results, these being the block of state updates and the anchor transaction.

The anchor transaction contains chain state transition, the AliasOutput and token transfers, so it must be signed.

**It is only possible to produce valid signatures of inputs of the anchor transaction by the quorum** of nodes. In this case, a confirmed anchor transaction becomes a cryptographical **proof of consensus** in the committee.  

To archive this, ISCP uses **BLS threshold signatures in combination with polynomial (Shamir) secret sharing** to identify the address controlling the chain state. In order for the secret keys to be distributed across the chain validators, a DKG (Distributed Key Generation) procedure is executed when starting a chain (using the Rabin-Gennaro algorithm).

## The Consensus Algorithm

The committee is of fixed size, thus we use a Byzantine Fault Tolerant (BFT) Algorithm, which guarantees consistency and byzantine fault tolerance if less than ⅓ of nodes are malicious.

As a basis for the ISCP consensus, the Asynchronous Common Subset (ACS) part of the HoneyBadgerBFT algorithm is used, with the exception of how the proposals are combined.

The rest of the consensus algorithm is built on top of the ACS. Each node supplies to the ACS its batch proposal which indicates a set of Request IDs, a timestamp, consensus and access mana pledge addresses, fee destination and a partial signature for generating non-forgeable entropy. Upon termination of the ACS, each honest node gets the same set of such proposals and aggregates them into the final batch in a deterministic way.

It is ensured that all honest nodes have the same input for the VM. After running the selected batch, the VM results are then collectively signed using the threshold signature. The signed transaction can be published by any node at this point. In order to minimize the load on the IOTA network, the nodes calculate a delay for posting the transaction to the network based on a deterministic permutation of the nodes relative to the local perception of time.

:::note
A more in-depth explanation of the topics described in this page can be found on the [architecture document](https://github.com/iotaledger/wasp/raw/master/documentation/ISCP%20architecture%20description%20v3.pdf)
:::
