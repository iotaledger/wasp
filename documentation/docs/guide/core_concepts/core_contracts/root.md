---
description: The root contract is the first smart contract deployed on the chain. It functions as a smart contract factory for the chain, and manages chain ownership and fees.
image: /img/logo/WASP_logo_dark.png
keywords:
- ISCP
- core
- root
- initialization
- entry points
- fees
- ownership
- Views
--- 
# The `root` Contract

The `root` contract is one of the [core contracts](overview.md) on each ISCP
chain.

The `root` contract provides the following functionalities:

- It is the first smart contract deployed on the chain. Upon receiving the `init` request, bootstraps the state of the chain. Part of the state initialization is the deployment of all other core contracts.

- It functions as a smart contract factory for the chain: upon request, it deploys other smart contracts and maintains an on-chain registry of smart contracts in its state.

- The contract registry keeps a list of contract records, which contain their respective name, hname, description and creator.

## Entry Points

The following are the functions/entry points of the `root` contract. Some of
them may require authorisation, i.e. can only be invoked by a specific caller,
for example the _chain owner_.

### init

The constructor. Automatically posted to the chain immediately after confirmation of the origin transaction, as the first call.
* Initializes base values of the chain according to parameters
* Sets the caller as the _chain owner_
* Sets chain fee color (default is _IOTA color_)
* Deploys all core contracts. The core contracts become part of the immutable state.
  It makes them callable just like any other smart contract deployed on the chain.

* **deployContract** - Deploys a smart contract on the chain, if the caller has
  deploy permission. Parameters:
    * hash of the _blob_ with the binary of the program and VM type
    * name of the instance. This is later used in the hashed form of _hname_
    * description of the instance

* Hash of the _blob_ with the binary of the program and VM type
* Name of the instance. This is later used in the hashed form of _hname_
* Description of the instance

### grantDeployPermission

### Views

### setContractFee

* **findContract** - Returns the record for a given smart contract (if it
  exists).

* **getContractRecords** - Returns the list of all smart contracts deployed on the chain and related records.
