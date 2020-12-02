// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package vmtypes

import (
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// SandboxView is an interface for read only call
type SandboxView interface {
	Params() codec.ImmutableCodec
	State() codec.ImmutableMustCodec
	MyBalances() coret.ColoredBalances
	// only calls view entry points
	Call(contractHname coret.Hname, entryPoint coret.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error)

	MyContractID() coret.ContractID
	Event(msg string)
	Eventf(format string, args ...interface{})
}
