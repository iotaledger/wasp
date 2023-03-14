// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"math/big"
	"time"

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
	// Params returns the parameters of the current call
	Params() *Params
	// ChainID returns the chain ID
	ChainID() ChainID
	// ChainOwnerID returns the AgentID of the current owner of the chain
	ChainOwnerID() AgentID
	// Contract returns the Hname of the current contract in the context
	Contract() Hname
	// AccountID returns the agentID of the current contract (i.e. chainID + contract hname)
	AccountID() AgentID
	// Caller is the agentID of the caller.
	Caller() AgentID
	// Timestamp returns the Unix timestamp of the current state in seconds
	Timestamp() time.Time
	// Log returns a logger that outputs on the local machine. It includes Panicf method
	Log() LogInterface
	// Utils provides access to common necessary functionality
	Utils() Utils
	// Gas returns sub-interface for gas related functions. It is stateful but does not modify chain's state
	Gas() Gas
	// GetNFTData returns information about a NFTID (issuer and metadata)
	GetNFTData(nftID iotago.NFTID) *NFT
	// CallView calls another contract. Only calls view entry points
	CallView(contractHname Hname, entryPoint Hname, params dict.Dict) dict.Dict
	// StateR returns the immutable k/v store of the current call (in the context of the smart contract)
	StateR() kv.KVStoreReader
}

type Params struct {
	Dict dict.Dict
	KVDecoder
}

type Helpers interface {
	Requiref(cond bool, format string, args ...interface{})
	RequireNoError(err error, str ...string)
}

type Authorize interface {
	RequireCaller(agentID AgentID)
	RequireCallerAnyOf(agentID []AgentID)
	RequireCallerIsChainOwner()
}

type Balance interface {
	// BalanceBaseTokens returns number of base tokens in the balance of the smart contract
	BalanceBaseTokens() uint64
	// BalanceNativeToken returns number of native token or nil if it is empty
	BalanceNativeToken(iotago.NativeTokenID) *big.Int
	// BalanceNativeTokens returns all native tokens owned by the smart contract
	BalanceNativeTokens() iotago.NativeTokens
	// OwnedNFTs returns the NFTIDs of NFTs owned by the smart contract
	OwnedNFTs() []iotago.NFTID
	// returns whether a given user owns a given amount of tokens
	HasInAccount(AgentID, *Assets) bool
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

	// Call calls the entry point of the contract with parameters and allowance.
	// If the entry point is full entry point, allowance tokens are available to be moved from the caller's
	// accounts (if enough). If the entry point is view, 'allowance' has no effect
	Call(target, entryPoint Hname, params dict.Dict, allowance *Assets) dict.Dict
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
	AllowanceAvailable() *Assets
	// TransferAllowedFunds moves assets from the caller's account to specified account within the budget set by Allowance.
	// Skipping 'assets' means transfer all Allowance().
	// The TransferAllowedFunds call mutates AllowanceAvailable
	// Returns remaining budget
	TransferAllowedFunds(target AgentID, transfer ...*Assets) *Assets
	// Send sends an on-ledger request (or a regular transaction to any L1 Address)
	Send(metadata RequestParameters)
	// EstimateRequiredStorageDeposit returns the amount of base tokens needed to cover for a given request's storage deposit
	EstimateRequiredStorageDeposit(r RequestParameters) uint64
	// StateAnchor properties of the anchor output
	StateAnchor() *StateAnchor
	// MintNFT mints an NFT
	// MintNFT(metadata []byte) // TODO returns a temporary ID

	// Privileged is a sub-interface of the sandbox which should not be called by VM plugins
	Privileged() Privileged
}

// Privileged is a sub-interface for core contracts. Should not be called by VM plugins
type Privileged interface {
	TryLoadContract(programHash hashing.HashValue) error
	CreateNewFoundry(scheme iotago.TokenScheme, metadata []byte) (uint32, uint64)
	DestroyFoundry(uint32) uint64
	ModifyFoundrySupply(serNum uint32, delta *big.Int) int64
	GasBurnEnable(enable bool)
	MustMoveBetweenAccounts(fromAgentID, toAgentID AgentID, assets *Assets)
	DebitFromAccount(AgentID, *Assets)
	CreditToAccount(AgentID, *Assets)

	SetBlockContext(bctx interface{})
	BlockContext() interface{}
}

// RequestParameters represents parameters of the on-ledger request. The output is build from these parameters
type RequestParameters struct {
	// TargetAddress is the target address. It may represent another chain or L1 address
	TargetAddress iotago.Address
	// Assets attached to the output, always taken from the caller's account.
	// It expected to contain base tokens at least the amount required for storage deposit
	// It depends on the context how it is handled when base tokens are not enough for storage deposit
	Assets *Assets
	// AdjustToMinimumStorageDeposit if true base tokens in attached fungible tokens will be added to meet minimum storage deposit requirements
	AdjustToMinimumStorageDeposit bool
	// Metadata is a request metadata. It may be nil if the output is just sending assets to L1 address
	Metadata *SendMetadata
	// SendOptions includes options of the output, such as time lock or expiry parameters
	Options SendOptions
}

type Gas interface {
	Burn(burnCode gas.BurnCode, par ...uint64)
	Budget() uint64
	Burned() uint64
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
	StateData            []byte
	Deposit              uint64
	NativeTokens         iotago.NativeTokens
}

type SendOptions struct {
	Timelock   time.Time
	Expiration *Expiration
}

type Expiration struct {
	Time          time.Time
	ReturnAddress iotago.Address
}

// SendMetadata represents content of the data payload of the output
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
	Hashing() Hashing
	ED25519() ED25519
	BLS() BLS
}

type Hashing interface {
	Blake2b(data []byte) hashing.HashValue
	Sha3(data []byte) hashing.HashValue
	Hname(name string) Hname
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
