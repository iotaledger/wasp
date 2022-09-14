package isc

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

const Million = uint64(1_000_000)

func DecodeOutputID(b []byte, def ...iotago.OutputID) (iotago.OutputID, error) {
	if len(b) != iotago.OutputIDLength {
		if len(def) == 0 {
			return iotago.OutputID{}, xerrors.Errorf("expected OutputID size %d, got %d bytes",
				iotago.OutputIDLength, len(b))
		}
		return def[0], nil
	}
	var ret iotago.OutputID
	copy(ret[:], b)
	return ret, nil
}

func EncodeOutputID(value iotago.OutputID) []byte {
	return value[:]
}

func UTXOInputFromMarshalUtil(mu *marshalutil.MarshalUtil) (*iotago.UTXOInput, error) {
	data, err := mu.ReadBytes(iotago.OutputIDLength)
	if err != nil {
		return nil, err
	}
	id, err := DecodeOutputID(data)
	if err != nil {
		return nil, err
	}
	return id.UTXOInput(), nil
}

func UTXOInputToMarshalUtil(id *iotago.UTXOInput, mu *marshalutil.MarshalUtil) {
	mu.WriteBytes(EncodeOutputID(id.ID()))
}
