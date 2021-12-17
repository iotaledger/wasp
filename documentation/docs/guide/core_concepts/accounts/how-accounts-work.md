---
keywords:
- Smart Contracts
- on-chain account
- Ownership
- Accounts Contract
description: IOTA Smart Contracts chains keeps a ledger of on-chain account balances.  ON-chain accounts are identified by an AgentID.
image: /img/tutorial/accounts.png
---

# How Accounts Work

IOTA Smart Contracts provide secure, trustless transfers of digitized assets:

- Between smart contracts on the same or different chains
- Between smart contracts and L1 addresses on the UTXO Ledger

On the UTXO Ledger, just like with any DLT, we have **trustless** and **atomic**
transfers of assets between addresses on the ledger. The tokens contained in the
address can be moved to another address by providing a valid signature using the
private key which controls the source address.

In IOTA Smart Contracts, the smart contracts which reside on chains are also owners of their
tokens. Each smart contract can receive tokens that are transferred to it and
can send tokens it controls to any other owner, be it another smart
contract, or an ordinary L1 address on the UTXO Ledger.

There are 2 types of entities that can control tokens:

* L1 addresses on the UTXO Ledger
* Smart contracts on IOTA Smart Contracts chains

There are 3 different types of trustless token transfers possible between those
entities. Each type involves a different mechanism of transfer:

* Between L1 address and smart contract
* Between smart contracts on the same chain
* Between smart contracts on different chains

To make the system homogenous, we introduce the following two concepts:

* `Agent ID`: Represents an owner of tokens independently of the type of
  owning entity.
* `On-chain account`: Represents the unit of ownership on the chain.

Each IOTA Smart Contracts chain keeps a ledger of on-chain account balances

## Account Ownership

### Smart Contract ID

Unlike with blockchain systems like Ethereum, we cannot simply represent the
smart contract by a blockchain address: IOTA Smart Contracts can have many blockchains. 
Each chain in IOTA Smart Contracts is identified by its _chain ID_. A chain can
contain many smart contracts on it. So, in IOTA Smart Contracts, each contract is identified by
an identifier that consists of the chain ID, and the _hname_ of the smart
contract. In human-readable form, the smart _contract ID_ looks like this:

```
cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::cebf5908
```

The part before `::` is the _chain ID_, the part after `::` is the _hname_ of
the smart contract.

### Agent ID

The agent ID is an identifier that generalizes and represents one of the two
agent types in one identifier: either an L1 address on the UTXO Ledger or a
smart contract ID.

It is easy to determine which one is represented by the particular agent ID: an
L1 address always has the _hname_ part all zero.

### Different Types of Account

Given that an on-chain account is identified by an AgentID:

- The AgentID for accounts owned by L1 entities (regular IOTA wallets) looks like:

```yaml
Hname: 0
Address: "some address"
```

- The AgentID for accounts owned by L2 entities (Smart Contracts):

```yaml
Hname: "Hname of the entity"
Address: "Address of the chain where the entity exists"
```

For example, the smart contract with hname `123` that exists on the chain with address `000`, can be identified on **any** chain by the following AgentID:

```yaml
Hname: 123
Address: 000
```

## The Accounts Contract

The `Accounts` contract manages which funds are owned by which accounts.

Internally there is a mapping of `Account (AgentID)` to `balances`, which can include normal IOTAs and/or any colored tokens.

By calling this contract it is possible to:

- [View current account balances](./view-account-balances.mdx)
- [Deposit funds to the chain](./how-to-deposit-to-a-chain.mdx)
- [Withdraw funds from the chain](./how-to-withdraw-from-a-chain.mdx)
- [Harvest](./the-common-account.mdx) - can only be called by the chain owner, to move funds from the chain common account to his account.

## Interoperability Between Chains

Each chain contains any number of accounts. Each account owns colored
tokens: a collection of `color: balance` pairs.

Each account on the chain is controlled by some `agent ID`. It means that tokens
contained in the account can only be moved by the entity behind that agent ID:

* If the _agent ID_ represents an address on the UTXO Ledger, the tokens can
  only be moved by a request signed by the private key of that address.
* If the _agent ID_ represents a smart contract, the tokens can only be moved by
  that smart contract. Note that the smart contract may reside on the same chain
  or another chain.

[![Example situation. There are two chains deployed, with three smart contracts and one address.](/img/tutorial/accounts.png)](/img/tutorial/accounts.png)

The picture illustrates an example situation. There are two chains deployed,
with respective IDs `Pmc7iH8b..` and `Klm314noP8..`. The pink chain `Pmc7iH8b..`
has two smart contracts on it (`3037` and `2225`) and the blue chain
`Klm314noP8..` has one smart contract (`7003`).

The UTXO Ledger has 1 address `P6ZxYXA2..`. The address `P6ZxYXA2..` controls
1337 iotas and 42 red tokens on the UTXO Ledger. The same address also controls
42 iotas on the pink chain and 8 green tokens on the blue chain. So, the owner
of the private key behind the address controls 3 different accounts: 1 on the L1
UTXO Ledger and 2 accounts on 2 different chains on L2.

At the same time, smart contract `7003` on the blue chain has 5 iotas on its
native chain and controls 11 iotas on the pink chain.

Note that “control over account” means the entity which has the private key can
move funds from it. For an ordinary address, it means the private key of the
address. For a smart contract, it means the private keys of the committee which
runs the chain (the smart contract program can only be executed with those
private keys).
