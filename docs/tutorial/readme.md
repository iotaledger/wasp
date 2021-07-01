# Exploring IOTA Smart Contracts

## Purpose

This document is an introductory tutorial of the IOTA Smart Contract Platform
(ISCP) for developers.

The level of this document is technical. Its target audience is software
engineers who want to understand the ISCP, and the direction it is taking, in
order to develop their own dApps and/or contribute to the development of the
ISCP and Wasp nodes.

The approach in this tutorial is to introduce main concepts through writing
unit tests for several example smart contracts. We use a Go testing package 
codenamed [_Solo_](../../packages/solo/readme.md) in all examples of the
tutorial.

Knowledge of Go programming and basic knowledge of Rust programming are 
prerequisites.

## Chapters

* [The _Solo_ package](01.md)
* [First example](01.md#first-example)
* [Tokens and the UTXO Ledger](02.md#tokens-and-the-utxo-ledger)
* [Creating a chain. Core contracts](03.md#creating-a-chain-core-contracts)
* [Writing and compiling first Rust smart contract](03.md#writing-and-compiling-first-rust-smart-contract)
* [Deploying and running a Rust smart contract](04.md#deploying-and-running-a-rust-smart-contract)
* [Structure of the smart contract](05.md#structure-of-the-smart-contract)
    * [State](05.md#state)
    * [Entry points](05.md#entry-points)
* [Panic. Exception handling](05.md#panic-exception-handling)
* [Invoking a smart contract](06.md)
    * [Sending a request](06.md)
    * [Calling a view](07.md)
* [Sending and receiving tokens by example](08.md)
    * [Sending and receiving tokens with the address](08.md#sending-and-receiving-tokens-with-the-address)
    * [Sending tokens to the smart contract](09.md#sending-tokens-to-the-smart-contract)
    * [Return of tokens in case of failure](10.md#return-of-tokens-in-case-of-failure)
    * [Sending tokens from smart contract to address](11.md) 
* [ISCP on-chain accounts. Controlling token balances](iscp_accounts.md)

## Annexes

* [`root` contract](root.md)
* [`_default` contract](_default.md)
* [`accounts` contract](accounts.md)
* [`blob` contract](blob.md)
* [`blocklog` contract](blocklog.md)
* [`eventlog` contract](eventlog.md)

