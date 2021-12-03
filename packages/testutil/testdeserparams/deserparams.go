package testdeserparams

import iotago "github.com/iotaledger/iota.go/v3"

const MinDustDeposit = 582

func DeSerializationParameters() *iotago.DeSerializationParameters {
	return &iotago.DeSerializationParameters{
		RentStructure:  RentStructure(),
		MinDustDeposit: MinDustDeposit,
	}
}

func RentStructure() *iotago.RentStructure {
	return &iotago.RentStructure{
		VByteCost:    1,
		VBFactorData: 1,
		VBFactorKey:  1,
	}
}
