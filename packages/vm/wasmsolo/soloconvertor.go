package wasmsolo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

// SoloConvertor converts ISCP data types to WasmLib data types
type SoloConvertor struct{}

func (cvt SoloConvertor) ScAddress(address ledgerstate.Address) wasmtypes.ScAddress {
	return wasmtypes.AddressFromBytes(address.Bytes())
}

func (cvt SoloConvertor) ScAgentID(agentID *iscp.AgentID) wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(agentID.Bytes())
}

func (cvt SoloConvertor) ScChainID(chainID *iscp.ChainID) wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(chainID.Bytes())
}

func (cvt SoloConvertor) ScColor(color colored.Color) wasmtypes.ScColor {
	return wasmtypes.ColorFromBytes(color.Bytes())
}

func (cvt SoloConvertor) ScHash(hash hashing.HashValue) wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(hash.Bytes())
}

func (cvt SoloConvertor) ScHname(hname iscp.Hname) wasmtypes.ScHname {
	return wasmtypes.HnameFromBytes(hname.Bytes())
}

func (cvt SoloConvertor) ScRequestID(requestID iscp.RequestID) wasmtypes.ScRequestID {
	return wasmtypes.RequestIDFromBytes(requestID.Bytes())
}
