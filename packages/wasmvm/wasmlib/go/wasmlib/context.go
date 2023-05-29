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

var _ ScFuncClientContext = new(ScFuncContext)

func (ctx ScFuncContext) ClientContract(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	return hContract
}

func (ctx ScFuncContext) Host() ScHost {
	return nil
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view sandbox interface
type ScViewContext struct {
	ScSandboxView
}

var (
	_ ScViewClientContext = new(ScFuncContext)
	_ ScViewClientContext = new(ScViewContext)
)

func (ctx ScViewContext) ClientContract(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	return hContract
}
