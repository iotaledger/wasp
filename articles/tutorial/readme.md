# Exploring IOTA Smart Contracts

## Purpose
The document is an introductory tutorial of the IOTA Smart Contract 
platform (ISCP) for developers. 

The level of the document is technical. The target audience is software engineers who want 
to understand ISCP and the direction it is taking. in order to develop their own dApps 
and/or contribute to the development of the ISCP and the Wasp node. 

The approach in this tutorial is an introduction to main concepts through writing
tests to example smart contracts. 
For this, we use Go testing package codenamed [_Solo_](../../packages/solo/readme.md) in all examples in the tutorial.

The knowledge of Go programming and basics of Rust programming is a prerequisite. 
For detailed introduction into the ISCP architecture [follow the link](https://docs.google.com/document/d/1zNJZMdetCzwiBC85K6gWbnzgdT1RXuZCLsTclKdrVqc/edit?usp=sharing).

Chapters:

* [The _Solo_ package](1.md#the-_solo_-package)
* [First example](1.md#first-example)
* [Tokens and the Value Tangle](2.md#tokens-and-the-value-tangle)
* [Creating a chain. Core contacts](3.md#creating-a-chain-core-contacts)
* [Writing and compiling first Rust smart contract](3.md#writing-and-compiling-first-rust-smart-contract)
* [Deploying and running Rust smart contract](4.md#deploying-and-running-rust-smart-contract)
* [Structure of the smart contract](5.md#structure-of-the-smart-contract)
    * [State](5.md#state)
    * [Entry points](5.md#entry-points)
* [Panic. Exception handling](5.md#panic-exception-handling)
