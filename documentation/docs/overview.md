---
keywords:
- ISCP
- Smart Contracts
- Core Concepts
- Table of Contents
description: IOTA Smart Contract Protocol Documentation Overview
image: /img/logo/WASP_logo_dark.png
---

# ISCP Documentation

Goal of the documentation: Give a higher level overview of what ISCP is and how it works for the average developer building applications for ISCP or looking to run a chain/node. Doesn’t need to contain every explicit detail but should contain everything you need to know to set up a smart contract chain and run and interact with smart contracts.

## ISCP Core Concepts

- [What Are Smart Contracts?](/docs/guide/core_concepts/smart-contracts)
- [What is ISCP?](/docs/guide/core_concepts/iscp)
- [ISCP Architecture Overview](/docs/guide/core_concepts/iscp-architecture)
- [Committees and Validators](/docs/guide/core_concepts/validators)
- [Consensus](/docs/guide/core_concepts/consensus)
- [State, Transitions and State Anchoring](/docs/guide/core_concepts/states)
- Accounts
    - [How Accounts Work](/docs/guide/core_concepts/accounts/how-accounts-work)
    - [How to Deposit to a Chain](/docs/guide/core_concepts/accounts/how-to-deposit-to-a-chain)
    - [How to Withdraw From a Chain](/docs/guide/core_concepts/accounts/how-to-withdraw-from-a-chain)
- Interacting With Smart Contracts
    - [On-ledger Requests](/docs/guide/core_concepts/smartcontract-interaction/on-ledger-requests)
    - [Off-ledger Requests](/docs/guide/core_concepts/smartcontract-interaction/on-ledger-requests)
- Types of VMs/Languages
    - [How ISCP Works As a Language/VM Agnostic Platform](/docs/guide/core_concepts/vm-types/iscp-vm)
    - Rust/Wasm Based Smart Contracts
        - [Why and What Does It Look Like?](/docs/guide/core_concepts/vm-types/rust-wasm)
        - [Pros and Cons](/docs/guide/core_concepts/vm-types/rust-wasm)
    - Solidity/EVM Based Smart Contracts
        - [Why and What Does It Look Like?](/docs/guide/core_concepts/vm-types/evm)
        - [Pros and Cons](/docs/guide/core_concepts/vm-types/evm)

## Running ISCP Chains and Nodes

- Setting up a chain
    - Requirements
    - Configuration
    - Adding nodes/validators
    - Testing if it works
- Running a node
    - Requirements
    - Configuration
    - Dashboard
    - Testing if works
- Chain management
    - Administrative tasks


## Rust/WASM based smart contracts

- Introduction
- Smart contract example
- Deploying a smart contract
- Tooling
    - Scaffolding tool
    - Testing with Solo
    - CLI
- Reference
    - Available sandbox methods
- Examples/Tutorials
    - Hello World
    - Calling a view function
    - Sending a request to a smart contract function
    - Interacting with layer 1 assets/Account contract
    - Cross chain communication

## EVM based smart contracts

- Introduction
- Limitations
    - Limited by EVM, no layer 1 or cross-chain interaction yet
- Smart contract example
- Deploying a smart contract
    - Core concept of how EVM is implemented in ISCP
    - Why you should use existing EVM tooling
- Tooling
    - CLI
    - Metamask configuration
    - Hardhat configuration
    - Web3/Ethers.js setup
- External EVM references
- Examples/Tutorials
