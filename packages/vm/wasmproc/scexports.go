// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

type ScExports struct {
	ScSandboxObject
	wc *WasmContext
}

func NewScExports(wc *WasmContext) *ScExports {
	return &ScExports{wc: wc}
}

func (o *ScExports) SetBytes(keyID, typeID int32, bytes []byte) {
	o.wc.host.SetExport(keyID, string(bytes))
}
