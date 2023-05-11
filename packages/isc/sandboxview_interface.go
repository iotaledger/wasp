// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

// SandboxView is an interface for read only call
type SandboxView interface {
	SandboxBase

	// Privileged is a sub-interface of the sandbox that should never be called by VM plugins
	Privileged() PrivilegedView
}

// PrivilegedView is a sub-interface for core contracts. Should not be called by VM plugins
type PrivilegedView interface {
	GasBurnEnable(enable bool)
}
