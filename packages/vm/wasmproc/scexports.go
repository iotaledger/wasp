// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import "github.com/iotaledger/wasp/packages/vm/wasmhost"

type ScExports struct {
	ScSandboxObject
	wasmHost *wasmhost.WasmHost
}

func NewScExports(wasmHost *wasmhost.WasmHost) *ScExports {
	return &ScExports{wasmHost: wasmHost}
}

func (o *ScExports) SetBytes(keyID, typeID int32, bytes []byte) {
	o.wasmHost.SetExport(keyID, string(bytes))
}
