# The `governance` contract

The `governance` contract is one of the [core contracts](overview.md) on each ISCP
chain.

The `governance` contract provides the following functionality:

- Defines the set of identities that constitute the state controller (entity that owns the state output via the chain Alias Address). Its possible to add/remove addresses from the stateController (thus rotating the committee of validators).
- Defines who is the chain owner (the L1 entity that owns the chain - initially whoever deployed it) - the chain owner can collect special fees, and customize some chain-specific parameters.
- Defines the fees for request execution, and other chain-specific parameters.

### Entry points

The following are the functions / entry points of the `governance`, they can only be invoked by the chain owner.

- **rotateStateController** - called when committee is about to be rotated to the new address. If it fails, nothing happens. If it succeeds, the next state transition will become a governance transition, thus updating the state controller in the chain's AliasOutput.
- **addAllowedStateControllerAddress** - adds an address to the list of identities that constitute the state controller, this change will only become effective once `rotateStateController` is called  
- **removeAllowedStateControllerAddress** - removes an address to the list of identities that constitute the state controller, this change will only become effective once `rotateStateController` is called
- **delegateChainOwnership** - sets a new owner for the chain. This change will only be effective once `claimChainOwnership` is called
- **claimChainOwnership** - claims the ownership of the chain if the caller matches the identity set in `delegateChainOwnership`
- **setContractFee** - sets the fee for a particular contract
- **setChainInfo** -  allows the following chain parameters to be set: `MaxBlobSize`, `MaxEventSize`, `MaxEventsPerRequest`, `OwnerFee`, `ValidatorFee`

### Views

Can be called directly. Calling a view does not modify the state of the smart
contract.

- **getAllowedStateControllerAddresses** - returns the list of allowed state controllers
- **getChainOwner** - returns the AgentID of the chain owner
- **getFeeInfo** - returns the fees for a given contract
- **getChainInfo** - returns the following chain parameters: `MaxBlobSize`, `MaxEventSize`, `MaxEventsPerRequest`, `OwnerFee`, `ValidatorFee`
