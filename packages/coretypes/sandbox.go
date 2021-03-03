// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// Sandbox is an interface given to the processor to access the VMContext
// and virtual state, transaction builder and request parameters through it.
type Sandbox interface {
	// ChainOwnerID AgentID of the current owner of the chain
	ChainOwnerID() AgentID
	// ContractCreator agentID which deployed contract
	ContractCreator() AgentID
	// ContractID is the ID of the current contract. Take chainID with ctx.ContractID().ChainID()
	ContractID() ContractID
	// Caller is the agentID of the caller.
	Caller() AgentID
	// Params of the current call
	Params() dict.Dict
	// State k/v store of the current call (in the context of the smart contract)
	State() kv.KVStore
	// DeployContract deploys contract on the same chain. 'initParams' are passed to the 'init' entry point
	DeployContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict) error
	// Call calls the entry point of the contract with parameters and transfer.
	// If the entry point is full entry point, transfer tokens are moved between caller's and
	// target contract's accounts (if enough). If the entry point is view, 'transfer' has no effect
	Call(target Hname, entryPoint Hname, params dict.Dict, transfer ColoredBalances) (dict.Dict, error)
	// RequestID of the request in the context of which is the current call
	RequestID() RequestID
	// MintedSupply is number of free minted tokens, i.e. minted tokens which are sent to addresses
	// other than chain address. It is a proof of how many tokens has been minted with the
	// color of the transaction (after un-coloring all request tokens)
	// It may be used in use-cases such as Token Registry.
	// The color of the supply can be extracted from the RequestID
	// It is read-only method, it returns same value for all requests and all calls in the context of the transaction
	MintedSupply() int64
	// GetTimestamp return current timestamp of the context
	GetTimestamp() int64
	// GetEntropy 32 random bytes based on the hash of the current state transaction
	GetEntropy() hashing.HashValue // 32 bytes of deterministic and unpredictably random data
	// Balances returns colored balances owned by the smart contract
	Balances() ColoredBalances
	// IncomingTransfer return colored balances transferred by the call. They are already accounted into the Balances()
	IncomingTransfer() ColoredBalances
	// Balance return number of tokens of specific color in the balance of the smart contract
	Balance(col balance.Color) int64
	// TransferToAddress send tokens to the L1 ledger address
	TransferToAddress(addr address.Address, transfer ColoredBalances) bool
	// PostRequest sends cross-chain request
	PostRequest(par PostRequestParams) bool
	// Log interface provides local logging on the machine. It also includes Panicf methods which logs and panics
	Log() LogInterface
	// Event publishes "vmmsg" message through Publisher on nanomsg. It also logs locally, but it is not the same thing
	Event(msg string)
	//
	Utils() Utils
}

// PostRequestParams is parameters of the PostRequest call
type PostRequestParams struct {
	TargetContractID ContractID
	EntryPoint       Hname
	TimeLock         uint32
	Params           dict.Dict
	Transfer         ColoredBalances
}
