package codec

import (
	"fmt"
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

//nolint:gocyclo,funlen
func Encode(v interface{}) []byte {
	switch vt := v.(type) {
	case bool:
		return EncodeBool(vt)
	case int: // default to int64
		return EncodeInt64(int64(vt))
	case int8:
		return EncodeInt8(vt)
	case int16:
		return EncodeInt16(vt)
	case int32:
		return EncodeInt32(vt)
	case int64:
		return EncodeInt64(vt)
	case uint8:
		return EncodeUint8(vt)
	case uint16:
		return EncodeUint16(vt)
	case uint32:
		return EncodeUint32(vt)
	case uint64:
		return EncodeUint64(vt)
	case string:
		return EncodeString(vt)
	case *big.Int:
		return EncodeBigIntAbs(vt)
	case []byte:
		return vt
	case *hashing.HashValue:
		return EncodeHashValue(*vt)
	case hashing.HashValue:
		return EncodeHashValue(vt)
	case iotago.Address:
		return EncodeAddress(vt)
	case *isc.ChainID:
		return EncodeChainID(*vt)
	case isc.ChainID:
		return EncodeChainID(vt)
	case isc.AgentID:
		return EncodeAgentID(vt)
	case isc.RequestID:
		return EncodeRequestID(vt)
	case *isc.RequestID:
		return EncodeRequestID(*vt)
	case isc.Hname:
		return vt.Bytes()
	case isc.VMErrorCode:
		return vt.Bytes()
	case time.Time:
		return EncodeTime(vt)
	case util.Ratio32:
		return EncodeRatio32(vt)
	case *util.Ratio32:
		return EncodeRatio32(*vt)
	default:
		panic(fmt.Sprintf("Can't encode value %v", v))
	}
}
