package parameters

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"
)

// Parameters needed for interaction with iota.go

// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md

func DeSerializationParametersForTesting() *iotago.DeSerializationParameters {
	return testdeserparams.DeSerializationParameters() // TODO
}

// L1 describes parameters coming from the L1 node
type L1 struct {
	NetworkID                 uint64
	MaxTransactionSize        int
	DeSerializationParameters *iotago.DeSerializationParameters
}

func (l1 *L1) RentStructure() *iotago.RentStructure {
	return l1.DeSerializationParameters.RentStructure
}

func L1ForTesting() *L1 {
	return &L1{
		NetworkID:                 0,
		MaxTransactionSize:        32000,
		DeSerializationParameters: DeSerializationParametersForTesting(),
	}
}
