// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package vmtypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	// ChainID of the current chain
	ChainID() coretypes.ChainID
	// ChainOwnerID AgentID of the current owner of the chain
	ChainOwnerID() coretypes.AgentID
	// State is base level interface to access the key/value pairs in the virtual state
	State() codec.MutableMustCodec
	// Accounts provide access to on-chain accounts and to the current transfer
	Accounts() Accounts
	// Params access to parameters of the call
	Params() codec.ImmutableCodec
	// Caller is the agentID of the caller of the SC function
	Caller() coretypes.AgentID
	// MyContractID is the ID of the current contract
	MyContractID() coretypes.ContractID
	// MyAgentID is the AgentID representation of the MyContractID
	MyAgentID() coretypes.AgentID

	// CreateContract deploys contract on the same chain. 'initParams' are passed to the 'init' entry point
	CreateContract(programHash hashing.HashValue, name string, description string, initParams codec.ImmutableCodec) error
	// Call calls the entry point of the contract with parameters and transfer.
	// If the entry point is full entry point, transfer tokens are moved between caller's and target contract's accounts (if enough)
	// If the entry point if view, 'transfer' has no effect
	Call(target coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error)
	// ID of the request in the context of which is the current call
	RequestID() coretypes.RequestID
	// current timestamp
	GetTimestamp() int64
	// entropy base on the hash of the current state transaction
	GetEntropy() hashing.HashValue // 32 bytes of deterministic and unpredictably random data

	// TransferToAddress send tokens to ledger address (not contract)
	TransferToAddress(addr address.Address, transfer coretypes.ColoredBalances) bool
	// TransferCrossChain send funds to the targetAgentID account cross chain
	// to move own funds to own account use MyAgentID() as a targetAgentID
	TransferCrossChain(targetAgentID coretypes.AgentID, targetChainID coretypes.ChainID, transfer coretypes.ColoredBalances) bool

	// PostRequest sends cross chain request
	PostRequest(par NewRequestParams) bool
	// PostRequestToSelf send cross chain request to the caller contract on the same chain
	PostRequestToSelf(entryPoint coretypes.Hname, args dict.Dict) bool
	// PostRequestToSelfWithDelay sends request to itself with timelock for some seconds after the current timestamp
	PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, deferForSec uint32) bool

	// for testing
	// Event and Eventf publish "vmmsg" message through Publisher on nanomsg
	Event(msg string)
	Eventf(format string, args ...interface{})
	// Same as panic(), but added as a Sandbox method to emphasize that it's ok to panic from a SC.
	// A panic will be recovered, and Rollback() will be automatically called after.
	Panic(v interface{})
	// clear all updates, restore same context as in the beginning of the VM call
	Rollback()
}

type NewRequestParams struct {
	TargetContractID coretypes.ContractID
	EntryPoint       coretypes.Hname
	Timelock         uint32
	Params           dict.Dict
	Transfer         coretypes.ColoredBalances
}

// Accounts is an interface to access all functions with tokens
// in the local context of the call to a smart contract
type Accounts interface {
	MyBalances() coretypes.ColoredBalances
	Incoming() coretypes.ColoredBalances
	Balance(col balance.Color) int64
	MoveBalance(target coretypes.AgentID, col balance.Color, amount int64) bool
}
