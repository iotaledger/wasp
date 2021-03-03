# The `accounts` contract

The `accounts` contract is one of 4 [core contracts](coresc.md) on each ISCP chain. 

The function of the `accounts` contract is to keep a consistent ledger of on-chain accounts
for the entities which controls them: L1 addresses and smart contracts.

It provides functions to deposit and withdraw funds, also gives the count of total assets deposited 
on the chain. Note that the ledger of accounts is consistently maintained by the VM behind scenes,the `accounts` 
core smart contract provides frontend of authorized access to those account by outside users.  

### Entry Points

* **deposit** moves tokens attached as a transfer to the target account on the chain. 
If parameter `agentID` is specified the target account is `agentID`. If not, tokens are deposited
to the account controlled by the caller (the latter makes sense only if it is a request, not a on-chain call).

* **withdrawToAddress** is only valid if requested by the address (not a smart contract). It sends all funds controlled
by the caller (an address) to that address on L1.

* **withdrawToChain** is only valid if requested by the smart contract (not an address) from another chain. 
It sends all funds controlled by the caller (a smart contract) to the account on the native chain belonging to the caller.

### Views

* **getBalance** return balances of colored tokens controlled by the `agentID` specified in the call parameters. 
It returns balances as dictionory of `color: amount` pairs.

* **getTotalAssets** returns total assets on the chain. It always is equal to the sum of all on-chain accounts

* **getAccounts** return list of all non-empty accounts in the chain as a list of `agentIDs`.  

