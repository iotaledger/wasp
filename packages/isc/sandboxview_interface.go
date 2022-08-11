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

	// Privileged is a sub-interface of the sandbox that should never be called by VM plugins
	Privileged() PrivilegedView
}

// PrivilegedView is a sub-interface for core contracts. Should not be called by VM plugins
type PrivilegedView interface {
	GasBurnEnable(enable bool)
}
