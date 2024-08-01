package codec

import (
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/sui-go/sui"
)

var Address = NewCodec(decodeAddress, encodeAddress)

func decodeAddress(b []byte) (*cryptolib.Address, error) {
	if len(b) == 0 {
		return nil, errors.New("invalid Address size")
	}
	return cryptolib.NewAddressFromBytes(b)
}

func encodeAddress(a *cryptolib.Address) []byte {
	return a.Bytes()
}

var TokenScheme = NewCodec(decodeTokenScheme, encodeTokenScheme)

func decodeTokenScheme(b []byte) (iotago.TokenScheme, error) {
	ts, err := iotago.TokenSchemeSelector(uint32(b[0]))
	if err != nil {
		return nil, err
	}
	_, err = ts.Deserialize(b, serializer.DeSeriModePerformValidation, nil)
	return ts, err
}

func encodeTokenScheme(o iotago.TokenScheme) []byte {
	return lo.Must(o.Serialize(serializer.DeSeriModeNoValidation, nil))
}

var CoinType = NewCodec(decodeCoinType, encodeCoinType)

func decodeCoinType(b []byte) (ret isc.CoinType, err error) {
	if len(b) != len(ret) {
		return ret, fmt.Errorf("%T: bytes length must be %d", ret, len(ret))
	}

	return isc.CoinTypeFromBytes(b)
}

func encodeCoinType(coinType isc.CoinType) []byte {
	return coinType.Bytes()
}

var ObjectID = NewCodec(decodeObjectID, encodeObjectID)
var NFTID = ObjectID

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
