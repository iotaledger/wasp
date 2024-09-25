package codec

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

var (
	AgentID       = NewCodecFromBCS[isc.AgentID]()
	ChainID       = NewCodecFromBCS[isc.ChainID]()
	VMErrorCode   = NewCodecFromBCS[isc.VMErrorCode]()
	HashValue     = NewCodecFromBCS[hashing.HashValue]()
	Hname         = NewCodecFromBCS[isc.Hname]()
	Ratio32       = NewCodecFromBCS[util.Ratio32]()
	RequestID     = NewCodecFromBCS[isc.RequestID]()
	CallArguments = NewCodecFromBCS[isc.CallArguments]()
	Assets        = NewCodecFromBCS[*isc.Assets]()
	CoinBalances  = NewCodecFromBCS[isc.CoinBalances]()
)
