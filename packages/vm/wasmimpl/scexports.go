// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmimpl

type ScExports struct {
	ScDict
}

func (o *ScExports) SetString(keyId int32, value string) {
	o.vm.SetExport(keyId, value)
}
