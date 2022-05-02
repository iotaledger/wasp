package testdeserparams

import iotago "github.com/iotaledger/iota.go/v3"

func ProtocolParameters() *iotago.ProtocolParameters {
	return &iotago.ProtocolParameters{
		RentStructure: *RentStructure(),
	}
}

func RentStructure() *iotago.RentStructure {
	return &iotago.RentStructure{
		VByteCost:    1,
		VBFactorData: 1,
		VBFactorKey:  1,
	}
}
