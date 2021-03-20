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
	GetEntryPoint(code Hname) VMProcessorEntryPoint // always should return valid entry point
	GetDescription() string
}

// VMProcessorEntryPoint is an abstract interface by which VM is called by passing
// the Sandbox interface
type VMProcessorEntryPoint interface {
	Call(ctx interface{}) (dict.Dict, error)
}

var (
	ErrInternalWrongTypeEntryPoint = xerrors.New("internal error: wrong type of the entry point")
	ErrWrongTypeEntryPoint         = xerrors.New("wrong type of the entry point")
)
