// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/kv/dict"
)

// package present processor interface. It must be implemented by VM

// VMProcessor is a abstract interface to the VM processor instance.
type VMProcessor interface {
	GetEntryPoint(code Hname) (VMProcessorEntryPoint, bool)
	GetDefaultEntryPoint() VMProcessorEntryPoint // return EP which can only be called with Sandbox
	GetDescription() string
}

// VMProcessorEntryPoint is an abstract interface by which VM is called by passing
// the Sandbox interface
type VMProcessorEntryPoint interface {
	Call(ctx interface{}) (dict.Dict, error)
	IsView() bool
}

var (
	ErrInternalWrongTypeEntryPoint = xerrors.New("internal error: wrong type of the entry point")
	ErrWrongTypeEntryPoint         = xerrors.New("wrong type of the entry point")
)
