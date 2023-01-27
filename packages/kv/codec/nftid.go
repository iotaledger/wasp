package codec

import (
	"errors"

	iotago "github.com/iotaledger/iota.go/v3"
)

func DecodeNFTID(b []byte, def ...iotago.NFTID) (iotago.NFTID, error) {
	if b == nil {
		if len(def) == 0 {
			return iotago.NFTID{}, errors.New("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if len(b) != iotago.NFTIDLength {
		return iotago.NFTID{}, errors.New("cannot decode NFTID: invalid length")
	}
	nftID := iotago.NFTID{}
	copy(nftID[:], b)
	return nftID, nil
}

func MustDecodeNFTID(b []byte) iotago.NFTID {
	r, err := DecodeNFTID(b)
	if err != nil {
		panic(err)
	}
	return r
}

func EncodeNFTID(nftID iotago.NFTID) []byte {
	return nftID[:]
}
