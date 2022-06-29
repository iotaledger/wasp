---
description: Smart Contracts can only interact with the world by using the Sandbox interface which provides limited and deterministic access to the state through a key/value storage abstraction.
image: /img/sandbox.png
keywords:
- smart contracts
- sandbox
- interface
- storage abstraction
- explanation
---

# Sandbox Interface

A smart contract's access to the world has to be restricted. Imagine a smart contract that would directly tap into a weather forecast website: as the weather changes, the result of the contract's execution will change as well. This smart contract is not deterministic, meaning that you cannot reproduce the result yourself to verify it, because the result for each execution could be different.

The access to the chain's state has to be curated, too. The owner of the chain and developers of individual smart contracts are not necessarily the same entity, and a single malicious contract could ruin the whole chain if not limited to its own space. Instead of working on the state as a whole, each smart contract can only modify a part of it.

The only way for smart contracts to access data is to use the Sandbox interface (which is deterministic). It provides them with their own internal state as a list of key/value pairs.

![Sandbox](/img/sandbox.png)

Besides reading and writing to the contract state, the Sandbox interface allows smart contracts to access:

- The [ID](./accounts/how-accounts-work#agent-id) of the contract.
- The details of the current request or view call.
- The current request allowance, and a way to claim said allowance.
- The balances owned by the contract.
- The ID of whoever had deployed the contract.
- The timestamp of the current block.
- Cryptographic utilities like hashing, signature verification, and so on.
- The [events](../schema/events.mdx) dispatch.
- Entropy, which emulates randomness in an unpredictable yet deterministic way.
- Logging, which is used for debugging in a test environment.

The Sandbox API available in "view calls" is slightly more limited than the one available in normal "execution calls". You can find below a more detailed API for the Sandbox.

---

## Common Sandbox Methods
(available in both "Execution Sandbox" and "View Sandbox")

```go
// AccountID returns the agentID of the current contract
	AccountID() AgentID
	// Params returns the parameters of the current call
	Params() *Params
	// ChainID returns the chain ID
	ChainID() *ChainID
	// ChainOwnerID returns the AgentID of the current owner of the chain
	ChainOwnerID() AgentID
	// Contract returns the Hname of the current contract in the context
	Contract() Hname
	// ContractAgentID returns the agentID of the contract (i.e. chainID + contract hname)
	ContractAgentID() AgentID
	// ContractCreator returns the agentID that deployed the contract
	ContractCreator() AgentID
	// Timestamp returns the Unix timestamp of the current state in seconds
	Timestamp() time.Time
	// Log returns a logger that outputs on the local machine. It includes Panicf method
	Log() LogInterface
	// Utils provides access to common necessary functionality
	Utils() Utils
	// Gas returns sub-interface for gas related functions. It is stateful but does not modify chain's state
	Gas() Gas
	// GetNFTInfo returns information about a NFTID (issuer and metadata)
	GetNFTData(nftID iotago.NFTID) NFT // TODO should this also return the owner of the NFT?
	// CallView calls another contract. Only calls view entry points
	CallView(contractHname Hname, entryPoint Hname, params dict.Dict) dict.Dict

STATE (on view sandbox is read-only)



type Helpers interface {
	Requiref(cond bool, format string, args ...interface{})
	RequireNoError(err error, str ...string)
}

type Balance interface {
	// BalanceIotas returns number of iotas in the balance of the smart contract
	BalanceIotas() uint64
	// BalanceNativeToken returns number of native token or nil if it is empty
	BalanceNativeToken(id *iotago.NativeTokenID) *big.Int
	// BalanceFungibleTokens returns all fungible tokens: iotas and native tokens
	BalanceFungibleTokens() *FungibleTokens
	// OwnedNFTs returns the NFTIDs of NFTs owned by the smart contract
	OwnedNFTs() []iotago.NFTID
}



```


## Execution Sandbox


```go
	// Request return the request in the context of which the smart contract is called
	Request() Calldata

	// Call calls the entry point of the contract with parameters and allowance.
	// If the entry point is full entry point, allowance tokens are moved between caller's and
	// target contract's accounts (if enough). If the entry point is view, 'allowance' has no effect
	Call(target, entryPoint Hname, params dict.Dict, allowance *Allowance) dict.Dict
	// Caller is the agentID of the caller.
	Caller() AgentID
	// DeployContract deploys contract on the same chain. 'initParams' are passed to the 'init' entry point
	DeployContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict)
	// Event emits an event
	Event(msg string)
	// RegisterError registers an error
	RegisterError(messageFormat string) *VMErrorTemplate
	// GetEntropy 32 random bytes based on the hash of the current state transaction
	GetEntropy() hashing.HashValue
	// AllowanceAvailable specifies max remaining (after transfers) budget of assets the smart contract can take
	// from the caller with TransferAllowedFunds. Nil means no allowance left (zero budget)
	// AllowanceAvailable MUTATES with each call to TransferAllowedFunds
	AllowanceAvailable() *Allowance
	// TransferAllowedFunds moves assets from the caller's account to specified account within the budget set by Allowance.
	// Skipping 'assets' means transfer all Allowance().
	// The TransferAllowedFunds call mutates AllowanceAvailable
	// Returns remaining budget
	// TransferAllowedFunds fails if target does not exist
	TransferAllowedFunds(target AgentID, transfer ...*Allowance) *Allowance
	// TransferAllowedFundsForceCreateTarget does not fail when target does not exist.
	// If it is a random target, funds may be inaccessible (less safe)
	TransferAllowedFundsForceCreateTarget(target AgentID, transfer ...*Allowance) *Allowance
	// Send sends an on-ledger request (or a regular transaction to any L1 Address)
	Send(metadata RequestParameters)
	// SendAsNFT sends an on-ledger request as an NFTOutput
	SendAsNFT(metadata RequestParameters, nftID iotago.NFTID)
	// EstimateRequiredDustDeposit returns the amount of iotas needed to cover for a given request's dust deposit
	EstimateRequiredDustDeposit(r RequestParameters) uint64
	// StateAnchor properties of the anchor output
	StateAnchor() *StateAnchor
```
