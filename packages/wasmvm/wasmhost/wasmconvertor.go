// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// WasmConvertor converts ISCP data types to WasmLib data types
type WasmConvertor struct{}

func (cvt WasmConvertor) IscpAddress(address *wasmtypes.ScAddress) ledgerstate.Address {
	ret, _, err := ledgerstate.AddressFromBytes(address.Bytes())
	if err != nil {
		panic(err)
	}
	return ret
}

func (cvt WasmConvertor) IscpAgentID(agentID *wasmtypes.ScAgentID) *iscp.AgentID {
	address := agentID.Address()
	hname := agentID.Hname()
	return iscp.NewAgentID(cvt.IscpAddress(&address), cvt.IscpHname(hname))
}

func (cvt WasmConvertor) IscpChainID(chainID *wasmtypes.ScChainID) *iscp.ChainID {
	ret, err := iscp.ChainIDFromBytes(chainID.Bytes())
	if err != nil {
		panic(err)
	}
	return ret
}

func (cvt WasmConvertor) IscpColor(color *wasmtypes.ScColor) colored.Color {
	ret, err := colored.ColorFromBytes(color.Bytes())
	if err != nil {
		panic(err)
	}
	return ret
}

func (cvt WasmConvertor) IscpHash(hash *wasmtypes.ScHash) hashing.HashValue {
	ret, err := hashing.HashValueFromBytes(hash.Bytes())
	if err != nil {
		panic(err)
	}
	return ret
}

func (cvt WasmConvertor) IscpHname(hname wasmtypes.ScHname) iscp.Hname {
	return iscp.Hname(hname)
}

func (cvt WasmConvertor) IscpRequestID(requestID *wasmtypes.ScRequestID) iscp.RequestID {
	ret, err := iscp.RequestIDFromBytes(requestID.Bytes())
	if err != nil {
		panic(err)
	}
	return ret
}

func (cvt WasmConvertor) ScAddress(address ledgerstate.Address) wasmtypes.ScAddress {
	return wasmtypes.AddressFromBytes(address.Bytes())
}

func (cvt WasmConvertor) ScAgentID(agentID *iscp.AgentID) wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(agentID.Bytes())
}

func (cvt WasmConvertor) ScChainID(chainID *iscp.ChainID) wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(chainID.Bytes())
}

func (cvt WasmConvertor) ScColor(color *colored.Color) wasmtypes.ScColor {
	return wasmtypes.ColorFromBytes(color.Bytes())
}

func (cvt WasmConvertor) ScHash(hash hashing.HashValue) wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(hash.Bytes())
}

func (cvt WasmConvertor) ScHname(hname iscp.Hname) wasmtypes.ScHname {
	return wasmtypes.ScHname(hname)
}

func (cvt WasmConvertor) ScRequestID(requestID iscp.RequestID) wasmtypes.ScRequestID {
	return wasmtypes.RequestIDFromBytes(requestID.Bytes())
}
