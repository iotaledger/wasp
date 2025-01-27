package main

import (
	"fmt"
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/nnikolash/wasp-types-exported/packages/hashing"
	"github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	"github.com/nnikolash/wasp-types-exported/packages/util"
	"github.com/samber/lo"
)

func Decode[Res any](b []byte) (Res, error) {
	var res interface{} = lo.Empty[Res]()
	var err error

	switch res.(type) {
	case bool:
		res, err = codec.DecodeBool(b)
	case int: // default to int64
		res, err = codec.DecodeInt64(b)
	case int8:
		res, err = codec.DecodeInt8(b)
	case int16:
		res, err = codec.DecodeInt16(b)
	case int32:
		res, err = codec.DecodeInt32(b)
	case int64:
		res, err = codec.DecodeInt64(b)
	case uint8:
		res, err = codec.DecodeUint8(b)
	case uint16:
		res, err = codec.DecodeUint16(b)
	case uint32:
		res, err = codec.DecodeUint32(b)
	case uint64:
		res, err = codec.DecodeUint64(b)
	case string:
		res, err = codec.DecodeString(b)
	case *big.Int:
		res, err = codec.DecodeBigIntAbs(b)
	case []byte:
		res = b
	case *hashing.HashValue:
		res, err = codec.DecodeHashValue(b)
	case hashing.HashValue:
		res, err = codec.DecodeHashValue(b)
	case iotago.Address:
		res, err = codec.DecodeAddress(b)
	case *isc.ChainID:
		res, err = codec.DecodeChainID(b)
	case isc.ChainID:
		res, err = codec.DecodeChainID(b)
	case isc.AgentID:
		res, err = codec.DecodeAgentID(b)
	case isc.RequestID:
		res, err = codec.DecodeRequestID(b)
	case *isc.RequestID:
		res, err = codec.DecodeRequestID(b)
	case isc.Hname:
		res, err = codec.DecodeHname(b)
	case iotago.NFTID:
		res, err = codec.DecodeNFTID(b)
	case isc.VMErrorCode:
		res, err = codec.DecodeVMErrorCode(b)
	case time.Time:
		res, err = codec.DecodeTime(b)
	case util.Ratio32:
		res, err = codec.DecodeRatio32(b)
	case *util.Ratio32:
		res, err = codec.DecodeRatio32(b)
	default:
		panic(fmt.Sprintf("Attempt to decode unexpected type %T: value = %x", res, b))
	}

	if err != nil {
		return res.(Res), fmt.Errorf("failed to decode value %x as %T: %w", b, res, err)
	}

	return res.(Res), nil
}
