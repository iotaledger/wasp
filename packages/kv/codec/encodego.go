package codec

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// TODO: remove this function, use bcs.Encode() or similar instead
//
//nolint:gocyclo,funlen
func Encode(v any) []byte {
	switch vt := v.(type) {
	case bool:
		return Bool.Encode(vt)
	case int: // default to int64
		return Int64.Encode(int64(vt))
	case int8:
		return Int8.Encode(vt)
	case int16:
		return Int16.Encode(vt)
	case int32:
		return Int32.Encode(vt)
	case int64:
		return Int64.Encode(vt)
	case uint8:
		return Uint8.Encode(vt)
	case uint16:
		return Uint16.Encode(vt)
	case uint32:
		return Uint32.Encode(vt)
	case uint64:
		return Uint64.Encode(vt)
	case string:
		return String.Encode(vt)
	case *big.Int:
		return BigIntAbs.Encode(vt)
	case []byte:
		return vt
	case *hashing.HashValue:
		return HashValue.Encode(*vt)
	case hashing.HashValue:
		return HashValue.Encode(vt)
	case *cryptolib.Address:
		return Address.Encode(vt)
	case *isc.ChainID:
		return ChainID.Encode(*vt)
	case isc.ChainID:
		return ChainID.Encode(vt)
	case isc.AgentID:
		return AgentID.Encode(vt)
	case isc.RequestID:
		return RequestID.Encode(vt)
	case *isc.RequestID:
		return RequestID.Encode(*vt)
	case isc.Hname:
		return vt.Bytes()
	case sui.ObjectID:
		return ObjectID.Encode(vt)
	case isc.VMErrorCode:
		return vt.Bytes()
	case time.Time:
		return Time.Encode(vt)
	case util.Ratio32:
		return Ratio32.Encode(vt)
	case *util.Ratio32:
		return Ratio32.Encode(*vt)
	case common.Address:
		return vt.Bytes()
	default:
		panic(fmt.Sprintf("Can't encode value %v", v))
	}
}
