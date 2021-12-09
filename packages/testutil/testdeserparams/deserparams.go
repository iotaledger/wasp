package testdeserparams

import iotago "github.com/iotaledger/iota.go/v3"

func DeSerializationParameters() *iotago.DeSerializationParameters {
	return &iotago.DeSerializationParameters{
		RentStructure: RentStructure(),
	}
}

func RentStructure() *iotago.RentStructure {
	return &iotago.RentStructure{
		VByteCost:    1,
		VBFactorData: 1,
		VBFactorKey:  1,
	}
}
