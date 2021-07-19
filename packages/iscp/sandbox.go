// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	// AccountID agentID of the current contract
	AccountID() *AgentID
	// Params of the current call
	Params() dict.Dict
	// State k/v store of the current call (in the context of the smart contract)
	State() kv.KVStore
	// RequestID of the request in the context of which is the current call
	RequestID() RequestID
	// Balance return number of tokens of specific color in the balance of the smart contract
	Balance(col ledgerstate.Color) uint64
	// Balances returns colored balances owned by the smart contract
	Balances() *ledgerstate.ColoredBalances
	// Call calls the entry point of the contract with parameters and transfer.
	// If the entry point is full entry point, transfer tokens are moved between caller's and
	// target contract's accounts (if enough). If the entry point is view, 'transfer' has no effect
	Call(target, entryPoint Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances) (dict.Dict, error)
	// Caller is the agentID of the caller.
	Caller() *AgentID
	// ChainID id of the chain
	ChainID() *ChainID
	// ChainOwnerID agentID of the current owner of the chain
	ChainOwnerID() *AgentID
	// Contract Hname of the contract in the current chain
	Contract() Hname
	// ContractCreator agentID which deployed contract
	ContractCreator() *AgentID
	// DeployContract deploys contract on the same chain. 'initParams' are passed to the 'init' entry point
	DeployContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict) error
	// Event publishes "vmmsg" message through Publisher on nanomsg. It also logs locally, but it is not the same thing
	Event(msg string)
	// GetEntropy 32 random bytes based on the hash of the current state transaction
	GetEntropy() hashing.HashValue // 32 bytes of deterministic and unpredictably random data
	// GetTimestamp return current timestamp of the context
	GetTimestamp() int64
	// IncomingTransfer return colored balances transferred by the call. They are already accounted into the Balances()
	IncomingTransfer() *ledgerstate.ColoredBalances
	// Log interface provides local logging on the machine. It also includes Panicf methods which logs and panics
	Log() LogInterface
	// Minted represents new colored tokens which has been minted in the request transaction
	// Note that the minted tokens can be sent to any addresses, not necessarily the chain address
	Minted() map[ledgerstate.Color]uint64
	// Send one generic method for sending assets with ledgerstate.ExtendedLockedOutput
	// replaces TransferToAddress and Post1Request
	Send(target ledgerstate.Address, tokens *ledgerstate.ColoredBalances, metadata *SendMetadata, options ...SendOptions) bool
	// Utils provides access to common necessary functionality
	Utils() Utils
	// Internal for use in native hardcoded contracts
	BlockContext(construct func(sandbox Sandbox) interface{}, onClose func(interface{})) interface{}
	// properties of the anchor output
	StateAnchor() StateAnchor
}

// properties of the anchor output/transaction in the current context
type StateAnchor interface {
	StateAddress() ledgerstate.Address
	GoverningAddress() ledgerstate.Address
	StateIndex() uint32
	StateHash() hashing.HashValue
	OutputID() ledgerstate.OutputID
}

type SendOptions struct {
	TimeLock         uint32 // unix seconds
	FallbackAddress  ledgerstate.Address
	FallbackDeadline uint32 // unix seconds
}

// RequestMetadata represents content of the data payload of the output
type SendMetadata struct {
	TargetContract Hname
	EntryPoint     Hname
	Args           dict.Dict
}
