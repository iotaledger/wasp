// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// SandboxView is an interface for read only call
type SandboxView interface {
	// AccountID agentID of the current contract
	AccountID() *AgentID
	// Params of the current call
	Params() dict.Dict
	// State immutable k/v store of the current call (in the context of the smart contract)
	State() kv.KVStoreReader
	// Balances is colored balances owned by the contract
	Balances() colored.Balances
	// Call calls another contract. Only calls view entry points
	Call(contractHname Hname, entryPoint Hname, params dict.Dict) (dict.Dict, error)
	// ChainID is the chain
	ChainID() *ChainID
	// ChainOwnerID AgentID of the current owner of the chain
	ChainOwnerID() *AgentID
	// Contract ID
	Contract() Hname
	// ContractCreator agentID which deployed contract
	ContractCreator() *AgentID
	// GetTimestamp return timestamp of the current state
	GetTimestamp() int64
	// Log interface provides local logging on the machine. It includes Panicf method
	Log() LogInterface
	// Utils provides access to common necessary functionality
	Utils() Utils
}
