# The `root` contract

The `root` contract is one of the [core contracts](overview.md) on each ISCP
chain.

The `root` contract provides the following functions:

- It is the first smart contract deployed on the chain. Upon receiving the `init` request, bootstraps the state of the chain. Part of the state initialization is the deployment of all other core contracts.

- It functions as a smart contract factory for the chain: upon request, it deploys other smart contracts and maintains an on-chain registry of smart contracts in its state.

- The contract registry keeps a list of contract records, which contain their respective name, hname, description and creator.

### Entry points

The following are the functions / entry points of the `root` contract. Some of
them may require authorisation, i.e. can only be invoked by a specific caller,
for example the _chain owner_.

* **init** - The constructor. Automatically posted to the chain immediately after
  confirmation of the origin transaction, as the first call.
    * Initializes base values of the chain according to parameters
    * sets the caller as the _chain owner_
    * sets chain fee color (default is _IOTA color_)
    * deploys all core contracts. The core contracts become part of the immutable state.
      It makes them callable just like any other smart contract deployed on the chain.

* **deployContract** - Deploys a smart contract on the chain, if the caller has
  deploy permission. Parameters:
    * hash of the _blob_ with the binary of the program and VM type
    * name of the instance. This is later used in the hashed form of _hname_
    * description of the instance

* **grantDeployPermission** - Chain owner grants deploy permission to an agent ID

* **revokeDeployPermission** - Chain owner revokes deploy permission from an agent ID

### Views

Can be called directly. Calling a view does not modify the state of the smart
contract.

* **findContract** - Returns the record for a given smart contract (if it
  exists).

* **getContractRecords** - Returns the list of all smart contracts deployed on the chain and related records.
