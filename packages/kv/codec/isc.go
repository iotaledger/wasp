package codec

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

var (
	AgentID     = NewCodecEx(isc.AgentIDFromBytes)
	ChainID     = NewCodecEx(isc.ChainIDFromBytes)
	VMErrorCode = NewCodecEx(isc.VMErrorCodeFromBytes)
	HashValue   = NewCodecEx(hashing.HashValueFromBytes)
	Hname       = NewCodecEx(isc.HnameFromBytes)
	Ratio32     = NewCodecEx(util.Ratio32FromBytes)
	RequestID   = NewCodecEx(isc.RequestIDFromBytes)
)
