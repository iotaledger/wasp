package codec

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
)

var (
	Address   = NewCodecFromBCS[*cryptolib.Address]()
	CoinType  = NewCodecFromBCS[coin.Type]()
	CoinValue = NewCodecFromBCS[coin.Value]()
	ObjectID  = NewCodecFromBCS[sui.ObjectID]()
	NFTID     = ObjectID
)
