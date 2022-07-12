---
description: The `governance` contract defines the set of identities that constitute the state controller, access nodes, who is the chain owner and the fees for request execution.  
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
- reference
--- 

# The `governance` Contract

The `governance` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts
chain.

The `governance` contract provides the following functionalities:

- It defines the set of identities that constitute the state controller (entity that owns the state output via the chain Alias Address). It is possible to add/remove addresses from the stateController (thus rotating the committee of validators).
- It defines who is the chain owner (the L1 entity that owns the chain - initially whoever deployed it). The chain owner can collect special fees, and customize some chain-specific parameters.
- It defines who are the entities allowed to have an access node.
- It defines the fee policy for the chain (gas price, what token in used to pay for gas, and the validator fee share).

---

## Entry Points

The following are the functions/entry points of the `governance` contract. Most governance entry points can only be invoked by the chain owner.

### `rotateStateController(S StateControllerAddress)`

Called when committee is about to be rotated to the new address `S`. If it fails, nothing happens. If it succeeds, the next state transition will become a governance transition, thus updating the state controller in the chain's AliasOutput.

Can only be invoked by the chain owner.

### `addAllowedStateControllerAddress(S StateControllerAddress)`

Adds the address `S` to the list of identities that constitute the state controller, this change will only become effective once `rotateStateController` is called.

Can only be invoked by the chain owner.

### `removeAllowedStateControllerAddress(S StateControllerAddress)`

Removes the address `S` from the list of identities that constitute the state controller, this change will only become effective once `rotateStateController` is called.

Can only be invoked by the chain owner.

### `delegateChainOwnership(o AgentID)`

Sets the AgentID `o` as the new owner for the chain. This change will only be effective once `claimChainOwnership` is called by `o`.

Can only be invoked by the chain owner.

### `claimChainOwnership()`

Claims the ownership of the chain if the caller matches the identity set in `delegateChainOwnership`.

### `setChainInfo(mb MaxBlobSize, me MaxEventSize, mr MaxEventsPerRequest)`

Allows the following chain parameters to be set by the chain owner: `MaxBlobSize`, `MaxEventSize`, `MaxEventsPerRequest`.

Can only be invoked by the chain owner.

### `setFeePolicy(g FeePolicy)`

Sets the fee policy for the chain. This includes the Token ID to be used for gas (by default is `null` for the base token), the GasPerToken (how many units of gas a token pays for), and the  ValidatorFee which is a value of 0-100, that means how much of the gas fees are distributed to the validators.

Can only be invoked by the chain owner.

### `addCandidateNode(ip PubKey, ic Certificate, ia API, i ForCommittee)`

Adds a node to the list of candidates.
The required parameters are the following:

- `ip` The public key of the node to be added
- `ic` The Certficate is a signed binary containing both the node public key, and their L1 address.
- `ia` The API url for the node
- `i` boolean (default false) - whether the candidate node is being added to be part of the committee, or just an access node

Can only be invoked by the access node owner (verified via the Certificate field).

### `revokeAccessNode(ip PubKey, ic Certificate, ia API, i ForCommittee)`

Removes a node from the list of candidates.
The parameters needed are the same as `addCandidateNode`

Can only be invoked by the access node owner (verified via the Certificate field).

### `changeAccessNodes(n actions)`

Iterates through a map of actions (`n`) and applies those actions. These actions are a map of pubKey -> to a byte value, this byte value can mean different things:

- 0 - Removes an access node from the access nodes list
- 1 - Accept a candidate node and adds it to the list of access nodes
- 2 - Drops an access node from the access nodes list and candidates list

Can only be invoked by the chain owner.

### `startMaintenance()`

Starts the maintenance mode.
Can only be invoked by the chain owner.

### `stopMaintenance()`

Stops the maintenance mode.
Can only be invoked by the chain owner.

---

## Views

Can be called directly. Calling a view does not modify the state of the smart contract.

### `getAllowedStateControllerAddresses()`

Returns the list of allowed state controllers.

### `getChainOwner()`

Returns the AgentID of the chain owner.

### `getChainInfo()`

Returns the following chain parameters: `MaxBlobSize`, `MaxEventSize`, `MaxEventsPerRequest`.

### `getFeePolicy()`

Returns the fee policy for the chain (gas price, what token in used to pay for gas, and the validator fee share).

### `getChainNodes()`

Returns a list of the current AccessNodes, and another list with the current candidates.

### `getMaintenanceStatus()`

Returns whether the chain is ongoing maintenance.
