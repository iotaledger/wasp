// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"strconv"
)

// The Arguments struct is used to gather all arguments for a smart
// contract function call and encode it into a deterministic byte array
type Arguments struct {
	Encoder
	args ArgMap
}

func (a *Arguments) Set(key string, val []byte) {
	if a.args == nil {
		a.args = make(ArgMap)
	}
	a.args.Set(key, val)
}

func (a *Arguments) IndexedKey(key string, index int) string {
	return key + "." + strconv.Itoa(index)
}

func (a *Arguments) Mandatory(key string) {
	if a.args != nil && a.args.Get(key) != nil {
		return
	}
	panic("missing mandatory " + key)
}
