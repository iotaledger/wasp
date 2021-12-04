---
description: The `governance` contract defines the set of identities that constitute the state controller, who is the chain owner and  the fees for request execution.  
image: /img/logo/WASP_logo_dark.png
keywords:
- core contracts
- governance
- state controller
- identities
- chain owner
- rotate
- remove
- claim
- add
- chain info
- fee info
--- 

# The `governance` Contract

The `governance` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts
chain.

The `governance` contract provides the following functionalities:

- It defines the set of identities that constitute the state controller (entity that owns the state output via the chain Alias Address). It is possible to add/remove addresses from the stateController (thus rotating the committee of validators).
- It defines who is the chain owner (the L1 entity that owns the chain - initially whoever deployed it). The chain owner can collect special fees, and customize some chain-specific parameters.
- It defines the fees for request execution, and other chain-specific parameters.

## Entry Points

The following are the functions/entry points of the `governance` contract. They can only be invoked by the chain owner.

### rotateStateController

Called when committee is about to be rotated to the new address. If it fails, nothing happens. If it succeeds, the next state transition will become a governance transition, thus updating the state controller in the chain's AliasOutput.

### addAllowedStateControllerAddress

Adds an address to the list of identities that constitute the state controller, this change will only become effective once `rotateStateController` is called  

### removeAllowedStateControllerAddress

Removes an address to the list of identities that constitute the state controller, this change will only become effective once `rotateStateController` is called

### delegateChainOwnership

Sets a new owner for the chain. This change will only be effective once `claimChainOwnership` is called

### claimChainOwnership

Claims the ownership of the chain if the caller matches the identity set in `delegateChainOwnership`

### setContractFee

Sets the fee for a particular contract

### setChainInfo

Allows the following chain parameters to be set: `MaxBlobSize`, `MaxEventSize`, `MaxEventsPerRequest`, `OwnerFee`, `ValidatorFee`

## Views

Can be called directly. Calling a view does not modify the state of the smart contract.

### getAllowedStateControllerAddresses

Returns the list of allowed state controllers.

### getChainOwner

Returns the AgentID of the chain owner.

### getFeeInfo

Returns the fees for a given contract.

### getChainInfo

Returns the following chain parameters: `MaxBlobSize`, `MaxEventSize`, `MaxEventsPerRequest`, `OwnerFee`, `ValidatorFee`.
