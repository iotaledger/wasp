## The `root` contract

The `root` contract is one of 4 [core contracts](coresc.md) on each ISCP chain. 
Functions of the `root` contract:

- it is the first smart contract deployed on the chain. It initializes the state of the chain.
The part of state initialization is deployment of all 4 core contracts.

- be a smart contract factory for the chain: deploy other smart contracts and maintain on-chain registry of smart contracts

- manage chain ownership. The _chain owner_ is a special `agentID` (address or another smart contract).
Initially the deployer of the chain becomes the _chain owner_. Certain function on the chain can only be performed
by the _chain owner_. That includes change of the chain ownership itself. 

- Managing default fees of the chain. There are two types of fees: _default chain owner fee_ and _default validator fees_. 
Initially both are set to 0. 

### Entry points
The following are the functions / entry points of the `root` contract. Some of them may require authorisation, i.e.
can only be invoked by specific caller, for example _chain owner_.  
 
* **init** the constructor. Automatically called immediately after deployment, as the first call.
   * Initializes base values of the chain according to parameters: chainID, chain color, chain address
   * sets _chain owner_ to the caller 
   * sets chain fee color (default is _IOTA color_)
   * deploys all 4 core contracts
   
* **deployContract** deploys smart contract on the chain, if the csaller has a permission. Parameters:
   * hash of the _blob_ with the binary of the program and VM type
   * name of the instance. Later it is used in the hashed form of _hname_
   * description of teh instance   

* **grantDeployPermission** chain owner grants deploy permission to the owner ID

* **revokeDeployPermission** chain owner revokes deploy permission for the owner ID
 
* **delegateChainOwnership** prepares a successor (an agent ID) of the owner of the chain. The ownership is not transferred until claimed.
   
* **claimChainOwnership** the successor can claim ownership if it was delegated. Chain ownership changes.    

* **setDefaultFee** sets chain-wide default fee values. There are two of them: `validatorFee` and `chainOwnerFee`. 
In the beginning both are 0. 

* **setContractFee** sets fee values for a particular smart contract. There are two values for each smart contract: 
`validatorFee` and `chainOwnerFee`. If the value is 0, it means the fee is taken from the corresponding 
default value on the chain level.

### Views
Can be called from outside of the chain. Calling a view does not modify state of the smart contact.

* **findContract** returns the data of the particular smart contract (if it exists) in marshalled binary form.

* **getChainInfo** returns main values of the chain, such as chainID, color, address. It also returns registry of 
smart contracts in marshalled binary form 

* **getFeeInfo** returns fee information for the particular smart contract: `validatorFee` and `chainOwnerFee`. 
It takes into account default values if specific values for the smart contract are not set.   
