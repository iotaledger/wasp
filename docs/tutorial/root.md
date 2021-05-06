## The `root` contract

The `root` contract is one of the [core contracts](coresc.md) on each ISCP
chain.

The `root` contract provides the following functions:

- It is the first smart contract deployed on the chain. It initializes the state
  of the chain. Part of the state initialization is deployment of all other core
  contracts.

- It functions as a smart contract factory for the chain: it deploys other smart
  contracts and maintains an on-chain registry of smart contracts.

- It manages chain ownership. The _chain owner_ is a special `agentID`
  (L1 address or another smart contract). Initially the deployer of the chain
  becomes the _chain owner_. Certain functions on the chain can only be
  performed by the _chain owner_. That includes changing the chain ownership
  itself.

- It manages default fees for the chain. There are two types of default fees:
  _chain owner fees_ and _validator fees_. Initially both are set to 0.

### Entry points

The following are the functions / entry points of the `root` contract. Some of
them may require authorisation, i.e. can only be invoked by a specific caller,
for example the _chain owner_.

* **init** - The constructor. Automatically called immediately after deployment,
  as the first call.
    * Initializes base values of the chain according to parameters: chainID,
      chain color, chain address
    * sets _chain owner_ to the caller
    * sets chain fee color (default is _IOTA color_)
    * deploys all 4 core contracts

* **claimChainOwnership** - The new chain owner can claim ownership if it was
  delegated. Chain ownership changes.

* **delegateChainOwnership** - Prepares a successor (an agent ID) to become the
  owner of the chain. The ownership is not transferred until claimed.

* **deployContract** - Deploys a smart contract on the chain, if the caller has
  deploy permission. Parameters:
    * hash of the _blob_ with the binary of the program and VM type
    * name of the instance. This is later used in the hashed form of _hname_
    * description of the instance

* **grantDeployPermission** - Chain owner grants deploy permission to an agent
  ID

* **revokeDeployPermission** - Chain owner revokes deploy permission from an
  agent ID

* **setContractFee** - Sets fee values for a particular smart contract. There
  are two values for each smart contract: `validatorFee` and `chainOwnerFee`. If
  a value is 0, it means the chain's default fee will be taken.

* **setDefaultFee** - Sets chain-wide default fee values. There are two of
  them: `validatorFee` and `chainOwnerFee`. Initially both are 0.

### Views

Can be called directly. Calling a view does not modify the state of the smart
contract.

* **findContract** - Returns the data of the provided smart contract (if it
  exists) in marshalled binary form.

* **getChainInfo** - Returns main values of the chain, such as chain ID, chain
  owner ID, and description. It also returns a registry of all smart contracts
  in marshalled binary form

* **getFeeInfo** - Returns fee information for the particular smart
  contract: `validatorFee` and `chainOwnerFee`. It takes into account default
  values if specific values for the smart contract are not set.   
