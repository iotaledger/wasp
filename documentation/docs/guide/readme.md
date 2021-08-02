# ISCP Documentation

Goal of the documentation: Give a higher level overview of what ISCP is and how it works for the average developer building applications for ISCP or looking to run a chain/node. Doesn’t need to contain every explicit detail but should contain everything you need to know to set up a smart contract chain and run and interact with smart contracts.

## ISCP Core concepts

- [What are smart contracts?](/docs/guide/core_concepts/smart-contracts)
- [What is ISCP?](/docs/guide/core_concepts/iscp)
- [ISCP Architecture overview](/docs/guide/core_concepts/iscp-architecture)
- Committees and validators, consensus
- State, transitions and state anchoring
- Accounts
    - How accounts work
    - How to deposit to a chain
    - How to withdraw from a chain
- Interacting with smart contracts
    - On-ledger requests
    - Off-ledger requests
- Types of VMs/Languages
    - How ISCP works as a language/VM agnostic platform
    - Rust/Wasm based smart contracts
        - Why and what does it look like?
        - Pros and cons
    - Solidity/EVM based smart contracts
        - Why and what does it look like?
        - Pros and cons

## Running ISCP chains and nodes

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
    - Redux configuration
    - Web3/Ethers.js setup
- External EVM references
- Examples/Tutorials
