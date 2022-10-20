// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type (
	ScFuncContextFunction func(ScFuncContext)
	ScViewContextFunction func(ScViewContext)

	ScHost interface {
		ExportName(index int32, name string)
		Sandbox(funcNr int32, params []byte) []byte
		StateDelete(key []byte)
		StateExists(key []byte) bool
		StateGet(key []byte) []byte
		StateSet(key, value []byte)
	}
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\
var (
	host  ScHost
	utils ScSandboxUtils
)

const hexDigits = "0123456789abcdef"

func init() {
	wasmtypes.Bech32Decode = utils.Bech32Decode
	wasmtypes.Bech32Encode = utils.Bech32Encode
	wasmtypes.NewScHname = utils.Hname

	wasmtypes.HexDecode = func(hex string) []byte {
		if !has0xPrefix(hex) {
			panic("hex string missing 0x prefix")
		}
		digits := len(hex) - 2
		if (digits & 1) != 0 {
			panic("odd hex string length")
		}
		buf := make([]byte, digits/2)
		for i := 0; i < digits; i += 2 {
			buf[i/2] = (hexer(hex[i+2]) << 4) | hexer(hex[i+3])
		}
		return buf
	}

	wasmtypes.HexEncode = func(buf []byte) string {
		bytes := len(buf)
		hex := make([]byte, bytes*2)
		for i, b := range buf {
			hex[i*2] = hexDigits[b>>4]
			hex[i*2+1] = hexDigits[b&0x0f]
		}
		return "0x" + string(hex)
	}
}

func has0xPrefix(s string) bool {
	return len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X')
}

func hexer(hexDigit byte) byte {
	if hexDigit >= '0' && hexDigit <= '9' {
		return hexDigit - '0'
	}
	if hexDigit >= 'a' && hexDigit <= 'f' {
		return hexDigit - 'a' + 10
	}
	if hexDigit >= 'A' && hexDigit <= 'F' {
		return hexDigit - 'A' + 10
	}
	panic("invalid hex digit")
}

func ConnectHost(h ScHost) ScHost {
	oldHost := host
	host = h
	return oldHost
}

func ExportName(index int32, name string) {
	host.ExportName(index, name)
}

func Sandbox(funcNr int32, params []byte) []byte {
	return host.Sandbox(funcNr, params)
}

func StateDelete(key []byte) {
	host.StateDelete(key)
}

func StateExists(key []byte) bool {
	return host.StateExists(key)
}

func StateGet(key []byte) []byte {
	return host.StateGet(key)
}

func StateSet(key, value []byte) {
	host.StateSet(key, value)
}
