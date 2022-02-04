// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// SandboxBase is the common interface of Sandbox and SandboxView
type SandboxBase interface {
	Helpers
	Balance
	// AccountID returns the agentID of the current contract
	AccountID() *AgentID
	// Params returns the parameters of the current call
	Params() dict.Dict
	// ChainID returns the chain ID
	ChainID() *ChainID
	// ChainOwnerID returns the AgentID of the current owner of the chain
	ChainOwnerID() *AgentID
	// Contract returns the Hname of the current contract in the context
	Contract() Hname
	// ContractCreator returns the agentID that deployed the contract
	ContractCreator() *AgentID
	// Timestamp returns the UnixNano timestamp of the current state
	Timestamp() int64
	// Log returns a logger that outputs on the local machine. It includes Panicf method
	Log() LogInterface
	// Utils provides access to common necessary functionality
	Utils() Utils
	// Gas returns sub-interface for gas related functions. It is stateful but does not modify chain's state
	Gas() Gas
}

type Helpers interface {
	Requiref(cond bool, format string, args ...interface{})
	RequireNoError(err error, str ...string)
}

type Authorize interface {
	RequireCaller(agentID *AgentID, str ...string)
	RequireCallerAnyOf(agentID []*AgentID, str ...string)
	RequireCallerIsChainOwner(str ...string)
}

type Balance interface {
	// BalanceIotas returns number of iotas in the balance of the smart contract
	BalanceIotas() uint64
	// BalanceNativeToken returns number of native token or nil if it is empty
	BalanceNativeToken(id *iotago.NativeTokenID) *big.Int
	// Assets returns all assets: iotas and native tokens
	Assets() *Assets
}

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	SandboxBase
	Authorize

	// State k/v store of the current call (in the context of the smart contract)
	State() kv.KVStore
	// Request return the request in the context of which the smart contract is called
	Request() Calldata

	// Call calls the entry point of the contract with parameters and transfer.
	// If the entry point is full entry point, transfer tokens are moved between caller's and
	// target contract's accounts (if enough). If the entry point is view, 'transfer' has no effect
	Call(target, entryPoint Hname, params dict.Dict, allowance *Assets) dict.Dict
	// Caller is the agentID of the caller.
	Caller() *AgentID
	// DeployContract deploys contract on the same chain. 'initParams' are passed to the 'init' entry point
	DeployContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict)
	// Event emits an event
	Event(msg string)
	// GetEntropy 32 random bytes based on the hash of the current state transaction
	GetEntropy() hashing.HashValue
	// AllowanceAvailable specifies max remaining (after transfers) budget of assets the smart contract can take
	// from the caller with TransferAllowedFunds. Nil means no allowance left (zero budget)
	// AllowanceAvailable MUTATES with each call to TransferAllowedFunds
	AllowanceAvailable() *Assets
	// TransferAllowedFunds moves assets from the caller's account to specified account within the budget set by Allowance.
	// Skipping 'assets' means transfer all Allowance().
	// The TransferAllowedFunds call mutates AllowanceAvailable
	// Returns remaining budget
	// TransferAllowedFunds fails if target does not exist
	TransferAllowedFunds(target *AgentID, assets ...*Assets) *Assets
	// TransferAllowedFundsForceCreateTarget does not fail when target does not exist.
	// If it is a random target, funds may be inaccessible (less safe)
	TransferAllowedFundsForceCreateTarget(target *AgentID, assets ...*Assets) *Assets
	// Send sends a on-ledger request (or a regular transaction to any L1 Address)
	Send(metadata RequestParameters)
	// EstimateRequiredDustDeposit returns the amount of iotas needed to cover for a given request's dust deposit
	EstimateRequiredDustDeposit(r RequestParameters) uint64
	// StateAnchor properties of the anchor output
	StateAnchor() *StateAnchor

	// Privileged is a sub-interface of the sandbox which should not be called by VM plugins
	Privileged() Privileged
}

// Privileged is a sub-interface for core contracts. Should not be called by VM plugins
type Privileged interface {
	TryLoadContract(programHash hashing.HashValue) error
	CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int, metadata []byte) (uint32, uint64)
	DestroyFoundry(uint32) uint64
	ModifyFoundrySupply(serNum uint32, delta *big.Int) int64
	BlockContext(construct func(sandbox Sandbox) interface{}, onClose func(interface{})) interface{}
}

// RequestParameters represents parameters of the on-ledger request. The output is build from these parameters
type RequestParameters struct {
	// TargetAddress is the target address. It may represent another chain or L1 address
	TargetAddress iotago.Address
	// Assets attached to the output, always taken from the caller's account.
	// It expected to contain iotas at least the amount required for dust deposit
	// It depends on the context how it is handled when iotas are not enough for dust deposit
	Assets *Assets
	// AdjustToMinimumDustDeposit if true iotas in assets will be added to meet minimum dust deposit requirements
	AdjustToMinimumDustDeposit bool
	// Metadata is a request metadata. It may be nil if the output is just sending assets to L1 address
	Metadata *SendMetadata
	// SendOptions includes options of the output, such as time lock or expiry parameters
	Options SendOptions
}

type Gas interface {
	Burn(burnCode gas.BurnCode, par ...uint64)
	Budget() uint64
}

// StateAnchor contains properties of the anchor output/transaction in the current context
type StateAnchor struct {
	ChainID              ChainID
	Sender               iotago.Address
	OutputID             iotago.OutputID
	IsOrigin             bool
	StateController      iotago.Address
	GovernanceController iotago.Address
	StateIndex           uint32
	StateData            StateData
	Deposit              uint64
	NativeTokens         iotago.NativeTokens
}

type SendOptions struct {
	Timelock   *TimeData
	Expiration *Expiration
}

type Expiration struct {
	TimeData
	ReturnAddress iotago.Address
}

// RequestMetadata represents content of the data payload of the output
type SendMetadata struct {
	TargetContract Hname
	EntryPoint     Hname
	Params         dict.Dict
	Allowance      *Assets
	GasBudget      uint64
}

// Utils implement various utilities which are faster on host side than on wasm VM
// Implement deterministic stateless computations
type Utils interface {
	Base58() Base58
	Hashing() Hashing
	ED25519() ED25519
	BLS() BLS
}

type Hashing interface {
	Blake2b(data []byte) hashing.HashValue
	Sha3(data []byte) hashing.HashValue
	Hname(name string) Hname
}

type Base58 interface {
	Decode(s string) ([]byte, error)
	Encode(data []byte) string
}

type ED25519 interface {
	ValidSignature(data []byte, pubKey []byte, signature []byte) bool
	AddressFromPublicKey(pubKey []byte) (iotago.Address, error)
}

type BLS interface {
	ValidSignature(data []byte, pubKey []byte, signature []byte) bool
	AddressFromPublicKey(pubKey []byte) (iotago.Address, error)
	AggregateBLSSignatures(pubKeysBin [][]byte, sigsBin [][]byte) ([]byte, []byte, error)
}
