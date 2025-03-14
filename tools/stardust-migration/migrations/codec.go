package migrations

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	"github.com/samber/lo"
)

// DEPRECATED (see utils.go)

type Deserializable interface {
	Read(src io.Reader) error
}

func DeserializeKey[OldKey any](b old_kv.Key) OldKey {
	switch interface{}(lo.Empty[OldKey]()).(type) {
	case []byte:
		panic(fmt.Errorf("old key is never expected to have []byte type - use old_kv.Key instead"))
	case old_kv.Key:
		return interface{}(b).(OldKey)
	default:
		var v OldKey
		_ = lo.Must(deserializeInto([]byte(b), &v))
		return v
	}
}

func DeserializeValue[OldValue any](b []byte) OldValue {
	switch interface{}(lo.Empty[OldValue]()).(type) {
	case []byte:
		return interface{}(b).(OldValue)
	case *old_kv.Key:
		panic(fmt.Errorf("old value is never expected to have old_kv.Key type - use []byte instead"))
	default:
		var v OldValue
		_ = lo.Must(deserializeInto(b, &v))
		return v
	}
}

// Attempts to first use Read method of Deserializable interface,
// and if fails - uses decodeInto for basic types decoding.
// If the type is raw bytes or old kv.Key - does nothing.
func deserializeInto[Dest any](b []byte, v *Dest) (*Dest, error) {
	f := func(de Deserializable) (*Dest, error) {
		r := bytes.NewReader(b)
		lo.Must0(de.Read(r))

		if r.Len() != 0 {
			leftovers := lo.Must(io.ReadAll(r))
			panic(fmt.Sprintf(
				"Leftover bytes after reading record of type %T: initialValue = %x, leftover = %x, leftoverLen = %v",
				v, b, leftovers, r.Len(),
			))
		}

		return v, nil
	}

	if deserializable, isDeserializible := interface{}(&v).(Deserializable); isDeserializible {
		return f(deserializable)
	}
	if deserializable, isDeserializible := interface{}(v).(Deserializable); isDeserializible {
		return f(deserializable)
	}

	decodeInto(b, v)

	return v, nil
}

// DecodeInto decodes basic types. If the type is raw bytes or old kv.Key - does nothing.
func decodeInto[Res any](b []byte, dest *Res) error {
	var res interface{} = lo.Empty[*Res]()
	var err error

	switch res.(type) {
	case *[]byte, *old_kv.Key:
		res = b
	case *kv.Key:
		panic("new kv.Key is not expected to appear in decoding of old bytes")
	case *bool:
		res, err = old_codec.DecodeBool(b)
	case *int: // default to int64
		res, err = old_codec.DecodeInt64(b)
	case *int8:
		res, err = old_codec.DecodeInt8(b)
	case *int16:
		res, err = old_codec.DecodeInt16(b)
	case *int32:
		res, err = old_codec.DecodeInt32(b)
	case *int64:
		res, err = old_codec.DecodeInt64(b)
	case *uint8:
		res, err = old_codec.DecodeUint8(b)
	case *uint16:
		res, err = old_codec.DecodeUint16(b)
	case *uint32:
		res, err = old_codec.DecodeUint32(b)
	case *uint64:
		res, err = old_codec.DecodeUint64(b)
	case *string:
		res, err = old_codec.DecodeString(b)
	case **big.Int:
		res, err = old_codec.DecodeBigIntAbs(b)
	case *old_iotago.NativeTokenID:
		res, err = old_codec.DecodeNativeTokenID(b)
	// case **old_hashing.HashValue:
	// 	res, err = old_codec.DecodeHashValue(b)
	// case *old_hashing.HashValue:
	// 	res, err = old_codec.DecodeHashValue(b)
	// case *old_iotago.Address:
	// 	res, err = old_codec.DecodeAddress(b)
	// case **old_isc.ChainID:
	// 	res, err = old_codec.DecodeChainID(b)
	// case *old_isc.ChainID:
	// 	res, err = old_codec.DecodeChainID(b)
	// case *old_isc.AgentID:
	// 	res, err = old_codec.DecodeAgentID(b)
	// case *old_isc.RequestID:
	// 	res, err = old_codec.DecodeRequestID(b)
	// case **old_isc.RequestID:
	// 	res, err = old_codec.DecodeRequestID(b)
	// case *old_isc.Hname:
	// 	res, err = old_codec.DecodeHname(b)
	// case *old_iotago.NFTID:
	// 	res, err = old_codec.DecodeNFTID(b)
	// case *old_isc.VMErrorCode:
	// 	res, err = old_codec.DecodeVMErrorCode(b)
	// case *time.Time:
	// 	res, err = old_codec.DecodeTime(b)
	// case *old_util.Ratio32:
	// 	res, err = old_codec.DecodeRatio32(b)
	// case **old_util.Ratio32:
	// 	res, err = old_codec.DecodeRatio32(b)
	default:
		panic(fmt.Sprintf("Attempt to decode unexpected type %T: value = %x", res, b))
	}

	if err != nil {
		return fmt.Errorf("failed to decode value %x as %T: %w", b, res, err)
	}

	*dest = res.(Res)

	return nil
}

// Serialize converts the key to bytes using new codec (which is BCS).
// If the value is already bytes or kv.Key - does nothing.
func SerializeKey[Key any](k Key) kv.Key {
	switch k := interface{}(k).(type) {
	case []byte:
		panic("new key is never expected to have []byte type - use kv.Key instead")
	case kv.Key:
		return k
	default:
		return kv.Key(codec.Encode(k))
	}
}

// Serialize converts the value to bytes using new codec (which is BCS).
// If the value is already bytes or kv.Key - does nothing.
func SerializeValue[Val any](v Val) []byte {
	switch v := interface{}(v).(type) {
	case []byte:
		return v
	case kv.Key:
		panic("new value is never expected to have kv.Key type - use []byte instead")
	default:
		return codec.Encode(v)
	}
}
