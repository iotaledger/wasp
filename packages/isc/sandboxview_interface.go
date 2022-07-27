// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"github.com/iotaledger/wasp/packages/kv"
)

// SandboxView is an interface for read only call
type SandboxView interface {
	SandboxBase

	// State immutable k/v store of the current call (in the context of the smart contract)
	State() kv.KVStoreReader
}
