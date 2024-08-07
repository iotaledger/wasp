package codec

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/sui-go/sui"
)

var (
	Address  = NewCodecEx(cryptolib.NewAddressFromBytes)
	CoinType = NewCodecEx(isc.CoinTypeFromBytes)
	ObjectID = NewCodec(decodeObjectID, encodeObjectID)
)

func decodeObjectID(b []byte) (ret sui.ObjectID, err error) {
	if len(b) != len(ret) {
		return ret, fmt.Errorf("%T: bytes length must be %d", ret, len(ret))
	}
	copy(ret[:], b)
	return ret, nil
}

func encodeObjectID(objectID sui.ObjectID) []byte {
	return objectID[:]
}
