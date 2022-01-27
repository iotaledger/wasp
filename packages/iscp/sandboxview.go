// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// SandboxView is an interface for read only call
type SandboxView interface {
	SandboxBase

	// State immutable k/v store of the current call (in the context of the smart contract)
	State() kv.KVStoreReader
	// Call calls another contract. Only calls view entry points
	Call(contractHname Hname, entryPoint Hname, params dict.Dict) dict.Dict
}
