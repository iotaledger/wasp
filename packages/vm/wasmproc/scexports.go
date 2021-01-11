// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

type ScExports struct {
	ScSandboxObject
}

func NewScExports(vm *wasmProcessor) *ScExports {
	o := &ScExports{}
	o.vm = vm
	return o
}

func (o *ScExports) SetString(keyId int32, value string) {
	o.vm.SetExport(keyId, value)
}
