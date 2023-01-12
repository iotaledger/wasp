// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type ScImmutableState struct{}

func (d ScImmutableState) Exists(key []byte) bool {
	return StateExists(key)
}

func (d ScImmutableState) Get(key []byte) []byte {
	return StateGet(key)
}

type ScState struct {
	ScImmutableState
}

func (d ScState) Bytes() []byte {
	panic("ScState.Bytes: cannot encode state")
}

var _ wasmtypes.IKvStore = ScState{}

func (d ScState) Delete(key []byte) {
	StateDelete(key)
}

func (d ScState) Immutable() ScImmutableState {
	return ScImmutableState{}
}

func (d ScState) Set(key, value []byte) {
	StateSet(key, value)
}

func NewStateProxy() wasmtypes.Proxy {
	return wasmtypes.NewProxy(new(ScState))
}
