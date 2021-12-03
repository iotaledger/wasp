package parameters

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"
)

// Global parameters for dust calculation
// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md
// TODO: make configurable

func DeSerializationParameters() *iotago.DeSerializationParameters {
	return testdeserparams.DeSerializationParameters() // TODO
}

func RentStructure() *iotago.RentStructure {
	return testdeserparams.RentStructure() // TODO
}
