# Exploring IOTA Smart Contracts

This document is an introductory tutorial of the IOTA Smart Contracts Platform
for developers.

The level of this document is technical. Its target audience is software
engineers who want to understand the IOTA Smart Contracts, and the direction it is taking, in
order to develop their own dApps and/or contribute to the development of the
IOTA Smart Contracts and Wasp nodes.

There are two ways of seeing the same thing, a smart contract: as a program of a state machine and as a distributed,
fault-tolerant and decentralized system. Both views are correct.

In this tutorial we look at the smart contract as a deterministic program, which read and updates its state, the ledger.
We will not discuss how to run smart contract on a chain, which, in turn, is run by a distributed network validator nodes.
The latter is transparent to the logic of the smart contract.

The approach in this tutorial is to introduce main concepts through writing
unit tests for several example smart contracts. We use a Go testing package 
codenamed [Solo](https://github.com/iotaledger/wasp/tree/master/packages/solo) in all examples of the
tutorial.

Knowledge of Go programming and basic knowledge of Rust programming are prerequisites.
