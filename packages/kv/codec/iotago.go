package codec

import (
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

var Address = NewCodec(decodeAddress, encodeAddress)

func decodeAddress(b []byte) (*cryptolib.Address, error) {
	if len(b) == 0 {
		return nil, errors.New("invalid Address size")
	}
	return cryptolib.NewAddressFromBytes(b), nil
}

func encodeAddress(a *cryptolib.Address) []byte {
	return a.Bytes()
}

var Output = NewCodec(decodeOutput, encodeOutput)

func decodeOutput(b []byte) (iotago.Output, error) {
	o, err := iotago.OutputSelector(uint32(b[0]))
	if err != nil {
		return nil, err
	}
	_, err = o.Deserialize(b, serializer.DeSeriModePerformValidation, nil)
	return o, err
}

func encodeOutput(o iotago.Output) []byte {
	return lo.Must(o.Serialize(serializer.DeSeriModeNoValidation, nil))
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

var NativeTokenID = NewCodec(decodeNativeTokenID, encodeNativeTokenID)

func decodeNativeTokenID(b []byte) (ret iotago.NativeTokenID, err error) {
	if len(b) != len(ret) {
		return ret, fmt.Errorf("%T: bytes length must be %d", ret, len(ret))
	}
	copy(ret[:], b)
	return ret, nil
}

func encodeNativeTokenID(nftID iotago.NativeTokenID) []byte {
	return nftID[:]
}

var NFTID = NewCodec(decodeNFTID, encodeNFTID)

func decodeNFTID(b []byte) (ret iotago.NFTID, err error) {
	if len(b) != len(ret) {
		return ret, fmt.Errorf("%T: bytes length must be %d", ret, len(ret))
	}
	copy(ret[:], b)
	return ret, nil
}

func encodeNFTID(nftID iotago.NFTID) []byte {
	return nftID[:]
}
