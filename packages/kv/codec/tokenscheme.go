package codec

import (
	"errors"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

func DecodeTokenScheme(b []byte, def ...iotago.TokenScheme) (iotago.TokenScheme, error) {
	if len(b) == 0 {
		if len(def) > 0 {
			return def[0], nil
		}
		return nil, errors.New("wrong data length")
	}
	ts, err := iotago.TokenSchemeSelector(uint32(b[0]))
	if err != nil {
		return nil, err
	}
	_, err = ts.Deserialize(b, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func MustDecodeTokenScheme(b []byte, def ...iotago.TokenScheme) iotago.TokenScheme {
	t, err := DecodeTokenScheme(b, def...)
	if err != nil {
		panic(err)
	}
	return t
}

func EncodeTokenScheme(value iotago.TokenScheme) []byte {
	ret, err := value.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}
	return ret
}
