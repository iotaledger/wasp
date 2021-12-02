package parameters

import iotago "github.com/iotaledger/iota.go/v3"

// Global parameters for dust calculation
// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md
// TODO: make configurable

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
