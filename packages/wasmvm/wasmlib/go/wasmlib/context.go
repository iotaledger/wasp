// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract func sandbox interface
type ScFuncContext struct {
	ScSandboxFunc
}

var _ ScFuncCallContext = &ScFuncContext{}

func (ctx ScFuncContext) Host() ScHost {
	return nil
}

func (ctx ScFuncContext) InitFuncCallContext() {
}

func (ctx ScFuncContext) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	return hContract
}

// TODO deprecated
func (ctx ScFuncContext) Incoming() ScBalances {
	return ctx.IncomingTransfer()
}

func (ctx ScFuncContext) TransferToAddress(address wasmtypes.ScAddress, transfer ScTransfers) {
	ctx.Send(address, transfer)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view sandbox interface
type ScViewContext struct {
	ScSandbox
}

var _ ScViewCallContext = &ScViewContext{}

func (ctx ScViewContext) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	return hContract
}
