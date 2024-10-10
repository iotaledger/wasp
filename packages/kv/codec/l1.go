package codec

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

var (
	Address   = NewCodecFromBCS[*cryptolib.Address]()
	CoinType  = NewCodecFromBCS[coin.Type]()
	CoinValue = NewCodecFromBCS[coin.Value]()
	ObjectID  = NewCodecFromBCS[iotago.ObjectID]()
	NFTID     = ObjectID
)
