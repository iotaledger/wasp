// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type InitHost struct {
	funcs []ScFuncContextFunction
	views []ScViewContextFunction
}

var _ ScHost = &InitHost{}

func NewInitHost() *InitHost {
	return &InitHost{}
}

func (h InitHost) AddFunc(f ScFuncContextFunction) []ScFuncContextFunction {
	if f != nil {
		h.funcs = append(h.funcs, f)
	}
	return h.funcs
}

func (h InitHost) AddView(v ScViewContextFunction) []ScViewContextFunction {
	if v != nil {
		h.views = append(h.views, v)
	}
	return h.views
}

func (h InitHost) ExportName(index int32, name string) {
	panic("implement me")
}

func (h InitHost) ExportWasmTag() {
	h.ExportName(-1, "WASM::GO::SOLO")
}

func (h InitHost) Sandbox(funcNr int32, params []byte) []byte {
	panic("implement me")
}

func (h InitHost) StateDelete(key []byte) {
	panic("implement me")
}

func (h InitHost) StateExists(key []byte) bool {
	panic("implement me")
}

func (h InitHost) StateGet(key []byte) []byte {
	panic("implement me")
}

func (h InitHost) StateSet(key, value []byte) {
	panic("implement me")
}
