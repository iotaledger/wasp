---
keywords:
- ISCP
- Smart Contracts
- on-chain account
- Ownership
- Accounts Contract
description: ISCP chains keeps a ledger of on-chain account balances.  ON-chain accounts are identified by an AgentID.
image: /img/logo/WASP_logo_dark.png
---

# How Accounts Work

Each ISCP chain keeps a ledger of on-chain account balances.

## Account Ownership

An on-chain account is identified by an AgentID.

- The AgentID for accounts owned by L1 entities (regular IOTA wallets) looks like the following:

    ```yaml
    Hname: 0
    Address: "some address"
    ```

- The AgentID for accounts owned by L2 entities (Smart Contracts) :

    ```yaml
    Hname: "Hname of the entity"
    Address: "Address of the chain where the entity exists"
    ```

    _example_: the smart contract with hname `123` that exists on the chain with address `000`, can be identified on **any** chain by the following AgentID:

    ```yaml
    Hname: 123
    Address: 000
    ```

## The Accounts Contract

The `Accounts` contract manages what funds are owned by which accounts.

Internally there is a mapping of `Account (AgentID)` to `balances`, which can include normal IOTAs and/or any colored tokens.

By calling this contract its possible to:

- [View current account balances](./view-account-balances.mdx)
- [Deposit funds to the chain](./how-to-deposit-to-a-chain.mdx)
- [Withdraw funds from the chain](./how-to-withdraw-from-a-chain.mdx)
- [Harvest](./the-common-account.mdx) - can only be called by the chain owner, to move funds from the chain common account to his own account.
