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
	Params() dict.Dict
	State() kv.KVStore
	MyBalances() coretypes.ColoredBalances
	// only calls view entry points
	Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params dict.Dict) (dict.Dict, error)

	MyContractID() coretypes.ContractID
	Event(msg string)
	Eventf(format string, args ...interface{})
}
