package rentstructure

import iotago "github.com/iotaledger/iota.go/v3"

// TODO temporary. Global parameters for dust calculation

func Get() *iotago.RentStructure {
	return &iotago.RentStructure{
		VByteCost:    1,
		VBFactorData: 1,
		VBFactorKey:  1,
	}
}
