# Exploring IOTA Smart Contracts

## Purpose
The document is an introductory tutorial of the IOTA Smart Contract 
platform (ISCP) for developers. 

The level of the document is technical. The target audience is software engineers who want 
to understand ISCP and the direction it is taking, in order to develop their own dApps 
and/or contribute to the development of the ISCP and the Wasp node. 

The approach in this tutorial is to introduce to main concepts through writing
unit tests of example smart contracts. 
We use Go testing package codenamed [_Solo_](../../packages/solo/readme.md) in all examples of the tutorial.

The knowledge of Go programming and basics of Rust programming is a prerequisite. 
For detailed introduction into the ISCP architecture [follow the link](https://docs.google.com/document/d/1zNJZMdetCzwiBC85K6gWbnzgdT1RXuZCLsTclKdrVqc/edit?usp=sharing).

Chapters:

* [The _Solo_ package](01.md)
* [First example](01.md#first-example)
* [Tokens and the Value Tangle](02.md#tokens-and-the-value-tangle)
* [Creating a chain. Core contacts](03.md#creating-a-chain-core-contacts)
* [Writing and compiling first Rust smart contract](03.md#writing-and-compiling-first-rust-smart-contract)
* [Deploying and running Rust smart contract](04.md#deploying-and-running-rust-smart-contract)
* [Structure of the smart contract](05.md#structure-of-the-smart-contract)
    * [State](05.md#state)
    * [Entry points](05.md#entry-points)
* [Panic. Exception handling](05.md#panic-exception-handling)
* [Invoking a smart contract](06.md)
    * [Sending a request](06.md)
    * [Calling a view](07.md)
* [Sending and receiving funds by example](08.md)
    * [Sending and receiving tokens with the address](08.md#sending-and-receiving-tokens-with-the-address)
    * [Sending tokens to the smart contract](09.md#sending-tokens-to-the-smart-contract)
    * [Return of tokens in case of failure](10.md#return-of-tokens-in-case-of-failure)
* [ISCP on-chain accounts. Controlling token balances](iscp_accounts.md)