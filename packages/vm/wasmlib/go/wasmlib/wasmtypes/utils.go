// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// wrapper for simplified use by hashtypes
func base58Encode(buf []byte) string {
	// TODO
	// return string(wasmlib.Sandbox(wasmstore.FnUtilsBase58Encode, buf))
	return hex(buf)
}

func hex(buf []byte) string {
	const hexa = "0123456789abcdef"
	digits := len(buf) * 2
	res := make([]byte, digits)
	for _, b := range buf {
		digits--
		res[digits] = hexa[b&0x0f]
		digits--
		res[digits] = hexa[b>>4]
	}
	return string(res)
}

func Panic(msg string) {
	panic(msg)
}
