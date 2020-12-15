// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package vmtypes

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// SandboxView is an interface for read only call
type SandboxView interface {
	// ChainID of the current chain
	ChainID() coretypes.ChainID
	// ContractID is the ID of the current contract
	ContractID() coretypes.ContractID
	// GetTimestamp return timestamp of the current state
	GetTimestamp() int64
	// Params of the current call
	Params() dict.Dict
	// State is access to k/v store of the current call (in the context of the smart contract)
	State() kv.KVStore
	// Balances is colored balances owned by the contract
	Balances() coretypes.ColoredBalances
	// Call calls another contract. Only calls view entry points
	Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params dict.Dict) (dict.Dict, error)

	// Log interface provides local logging on the machine
	Log() LogInterface
}
