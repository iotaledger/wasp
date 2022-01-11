// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

// The Results struct contains the results from a smart contract function call
type Results struct {
	Decoder
	res ResMap
}

func (r Results) Exists(key string) bool {
	return r.res.Get(key) != nil
}

func (r Results) ForEach(keyValue func(key []byte, val []byte)) {
	for key, val := range r.res {
		keyValue([]byte(key), val)
	}
}

func (r Results) Get(key string) []byte {
	return r.res.Get(key)
}
