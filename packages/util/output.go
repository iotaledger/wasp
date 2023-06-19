package util

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

func OutputFromBytes(data []byte) (iotago.Output, error) {
	outputType := data[0]
	output, err := iotago.OutputSelector(uint32(outputType))
	if err != nil {
		return nil, err
	}
	_, err = output.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	return output, nil
}
