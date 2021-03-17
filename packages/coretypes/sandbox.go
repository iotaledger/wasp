// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	// ChainOwnerID AgentID of the current owner of the chain
	ChainOwnerID() *AgentID
	// ContractCreator agentID which deployed contract
	ContractCreator() *AgentID
	// ContractID is the ID of the current contract. Take chainID with ctx.ContractID().ChainID()
	ContractID() *ContractID
	// Caller is the agentID of the caller.
	Caller() *AgentID
	// Params of the current call
	Params() dict.Dict
	// State k/v store of the current call (in the context of the smart contract)
	State() kv.KVStore
	// DeployContract deploys contract on the same chain. 'initParams' are passed to the 'init' entry point
	DeployContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict) error
	// Call calls the entry point of the contract with parameters and transfer.
	// If the entry point is full entry point, transfer tokens are moved between caller's and
	// target contract's accounts (if enough). If the entry point is view, 'transfer' has no effect
	Call(target, entryPoint Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances) (dict.Dict, error)
	// RequestID of the request in the context of which is the current call
	RequestID() ledgerstate.OutputID
	// GetTimestamp return current timestamp of the context
	GetTimestamp() int64
	// GetEntropy 32 random bytes based on the hash of the current state transaction
	GetEntropy() hashing.HashValue // 32 bytes of deterministic and unpredictably random data
	// Balances returns colored balances owned by the smart contract
	// Balances returns all colored balances at the disposition of the smart contract
	Balances() *ledgerstate.ColoredBalances
	// IncomingTransfer return colored balances transferred by the call. They are already accounted into the Balances()
	IncomingTransfer() *ledgerstate.ColoredBalances
	// MintedSupply is number of new tokens minted in the request and newly minted color
	Minted() (ledgerstate.Color, uint64)
	// TODO proofs of ownership and mint - special collection of methods
	// Balance return number of tokens of specific color in the balance of the smart contract
	Balance(col ledgerstate.Color) uint64

	// TransferToAddress send tokens to the L1 ledger address
	// Deprecated: use Send instead
	TransferToAddress(addr ledgerstate.Address, transfer *ledgerstate.ColoredBalances) bool
	// PostRequest sends cross-chain request
	// Deprecated: use Send instead
	PostRequest(par PostRequestParams) bool

	// Send one generic method for sending assets with ledgerstate.ExtendedLockedOutput
	// replaces TransferToAddress and PostRequest
	Send(target ledgerstate.Address, tokens *ledgerstate.ColoredBalances, metadata *SendMetadata, options ...SendOptions) bool

	// Log interface provides local logging on the machine. It also includes Panicf methods which logs and panics
	Log() LogInterface
	// Event publishes "vmmsg" message through Publisher on nanomsg. It also logs locally, but it is not the same thing
	Event(msg string)
	//
	Utils() Utils
}

// PostRequestParams is parameters of the PostRequest call
// Deprecated: use SendTransfer instead
type PostRequestParams struct {
	TargetContractID ContractID
	EntryPoint       Hname
	TimeLock         uint32 // unix seconds
	Params           dict.Dict
	Transfer         *ledgerstate.ColoredBalances
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
