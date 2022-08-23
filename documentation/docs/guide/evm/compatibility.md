---
description: Compatibility between the ISC EVM layer and existing Ethereum smart contracts and tooling.
image: /img/logo/WASP_logo_dark.png
keywords:
- smart contracts
- EVM
- Ethereum
- Solidity
- limitations
- compatibility
- fees
- reference
---
# EVM compatibility in IOTA Smart Contracts

EVM support in IOTA Smart Contracts is provided by the `evm` core contract. Its main purpose is to store the EVM state (account balances, state, code, etc.) and to provide a way to execute EVM code in order to manipulate the state. All of this is done on top of the ISC layer, which itself provides the rest of the machinery needed to run smart contracts (signed requests, blocks, state, proofs, etc).

However, The ISC EVM layer is also designed to be as compatible as possible with existing Ethereum tools like [MetaMask](https://metamask.io/), which assume that the EVM code will be run on an Ethereum blockchain, composed of Ethereum blocks containing Ethereum transactions. Unfortunately, since ISC works in a fundamentally different way, providing 100% compatibility is not possible. We do our best to somehow "fool" the Ethereum tools into thinking they are interfacing with a real Ethereum node, but some differences in behavior are inevitable.

Here are some of the most important properties and limitations of EVM support in IOTA Smart Contracts:

- The Wasp node provides a JSON-RPC service, which is the standard protocol used by Ethereum tools. Upon receiving a signed Ethereum transaction via JSON-RPC, the transaction is wrapped into an ISC off-ledger request. The sender of the request is the Ethereum address that signed the original transaction (e.g. the Metamask account).

- While ISC contracts are identified by an [hname](../core_concepts/smart-contract-anatomy.md), EVM contracts are identified by their Ethereum address.

- EVM contracts are not listed in the chain's [contract registry](../core_concepts/core_contracts/root.md).

- EVM contracts cannot be called via regular ISC requests; they can only be
  called through the JSON-RPC service.
  As a consequence, EVM contracts cannot receive on-ledger requests.

- The EVM state is stored in raw form (in contrast with an Ethereum blockchain, which stores the state in a Merkle tree — it would be inefficient to do that since it would be duplicating work done by the ISC layer).

- Any Ethereum transactions present in an ISC block are executed by the `evm` core contract, updating the EVM state accordingly. To provide compatibility with EVM tools, a "fake" Ethereum block is also created and stored. Not being part of a real Ethereum blockchain, some attributes of the blocks will contain dummy values (e.g. `stateRoot`, `nonce`, etc).

- Each stored block contains the executed Ethereum transactions and corresponding Ethereum receipts. If storage is limited, it is possible to configure EVM so that the latest N blocks are stored.

- There is no guaranteed *block time*. A new EVM "block" will be created only when an ISC block is created, and ISC does not enforce an average block time.

- Any Ethereum address is accepted as a valid `AgentID`, and thus can own L2 tokens on an IOTA Smart Contract chain, just like IOTA addresses.

- The Ethereum balance of an account is tied to its L2 ISC balance in the token used to pay for gas (e.g. by default `eth_getBalance` will return the L2 base token balance of the given Ethereum account). Any attempt to directly modify an Ethereum account balance will fail (e.g. attaching value to a transaction, calling `<address>.transfer(...)`, etc).

- In order to manipulate the owned ISC tokens (and in general, to access ISC functionality), there is a [special Ethereum contract](magic.md) that provides bindings to the ISC sandbox (e.g. call `isc.send(...)` to send tokens).

- The used EVM gas is converted to ISC gas before being charged to the sender. The conversion ratio is configurable. The token used to pay for gas is the same token configured in the ISC chain (IOTA by default). The gas fee is debited from the sender's L2 account, and it has to be deposited beforehand.
