package codec

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

func DecodeOutput(b []byte) (iotago.Output, error) {
	o, err := iotago.OutputSelector(uint32(b[0]))
	if err != nil {
		return nil, err
	}
	_, err = o.Deserialize(b, serializer.DeSeriModePerformValidation, nil)
	return o, err
}

func MustDecodeOutput(b []byte) iotago.Output {
	o, err := DecodeOutput(b)
	if err != nil {
		panic(err)
	}
	return o
}
