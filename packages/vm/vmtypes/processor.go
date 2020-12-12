// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package vmtypes

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// package present processor interface. It must be implemented by VM

// Processor is a abstract interface to the VM processor instance.
type Processor interface {
	GetEntryPoint(code coretypes.Hname) (EntryPoint, bool)
	GetDescription() string
}

// EntryPoint is an abstract interface by which VM is called by passing
// the Sandbox interface
type EntryPoint interface {
	WithGasLimit(int) EntryPoint
	Call(ctx Sandbox) (dict.Dict, error)
	IsView() bool
	CallView(ctx SandboxView) (dict.Dict, error)
}

var ErrWrongTypeEntryPoint = fmt.Errorf("wrong type of entry point")

// nilEntryPoint is the entry point implementation which does nothing when called
type nilEntryPoint bool

var (
	// NilEntryPointFull represents empty full entry point
	NilEntryPointFull = nilEntryPoint(false)
	// NilEntryPointView represents empty view entry point
	NilEntryPointView = nilEntryPoint(true)
)

func (n nilEntryPoint) WithGasLimit(_ int) EntryPoint {
	return n
}

func (n nilEntryPoint) Call(_ Sandbox) (dict.Dict, error) {
	if n {
		panic("wrong call of full entry point")
	}
	return nil, nil
}

func (n nilEntryPoint) IsView() bool {
	return bool(n)
}

func (n nilEntryPoint) CallView(_ SandboxView) (dict.Dict, error) {
	if !n {
		panic("wrong call of view entry point")
	}
	return nil, nil
}
