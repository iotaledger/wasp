![Wasp logo](WASP_logo_dark.png)
# Welcome to the Wasp repository! 

_Wasp_ is a [node software](https://github.com/iotaledger/wasp) developed by [IOTA Foundation](http://iota.org) 
to run the _IOTA Smart Contract Protocol_ (_ISC_, _ISCP_ in short) on top of the _IOTA Tangle_. 
Please find here a [high level introduction](https://blog.iota.org/an-introduction-to-iota-smart-contracts-16ea6f247936) 
into ISC. 

A _smart contract_ is a distributed software agent which keeps its state in the immutable ledger. 
The state is an append-only structure which evolves with each _request_ to the smart contract. 

State of the smart contract, including tokens deposited into it and the attached arbitrary data, 
is anchored in the _Value Tangle_, the [UTXO ledger](articles/intro/utxo.md). 
So, the IOTA ledger ensures state is immutable. 
 
Each SC is run by the distributed and leaderless _committee_ of Wasp nodes. 
The main purpose of the _committee_ is to ensure consistent transition from the previous state to the next, 
according to the attached program. The program itself is immutably stored with the smart contract too. 

So, IOTA smart contracts are run by the network of Wasp nodes, all connected to the Tangle.

The articles below explains how to run a Wasp node on the Pollen network, also 
concepts and architecture of ISCP and Wasp. 
We describe it using several PoC smart contracts as an example.

_Disclaimer. Wasp node and articles is a work in progress, and most likely will always be. 
The software presented in this repo is not ready for use in commercial settings or whenever processing 
of critical data is involved._  

## PoC smart contracts
- [Main concepts with _DonateWithFeedback_](articles/intro/dwf.md)
- [Deployment of the smart contract](articles/intro/deploy.md)
- [Handling tagged tokens with _TokenRegistry_ and _FairAuction_ smart contracts](articles/intro/tr-fa.md)
- [Short intoduction to UTXO ledger and digital assets](articles/intro/utxo.md)

## Instructions, docs
- [How to run a Wasp node](articles/docs/runwasp.md)
- [Wasp Pubisher](articles/docs/publisher.md)
