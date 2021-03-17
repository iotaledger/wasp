// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// SandboxView is an interface for read only call
type SandboxView interface {
	// ChainOwnerID AgentID of the current owner of the chain
	ChainOwnerID() *AgentID
	// ContractCreator agentID which deployed contract
	ContractCreator() *AgentID
	// ContractID is the ID of the current contract
	ContractID() *ContractID
	// GetTimestamp return timestamp of the current state
	GetTimestamp() int64
	// Params of the current call
	Params() dict.Dict
	// State immutable k/v store of the current call (in the context of the smart contract)
	State() kv.KVStoreReader
	// Call calls another contract. Only calls view entry points
	Call(contractHname Hname, entryPoint Hname, params dict.Dict) (dict.Dict, error)
	// Balances is colored balances owned by the contract
	Balances() *ledgerstate.ColoredBalances
	// Log interface provides local logging on the machine. It includes Panicf method
	Log() LogInterface
	//
	Utils() Utils
}
