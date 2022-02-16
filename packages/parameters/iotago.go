package parameters

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"
)

// Parameters needed for interaction with iota.go
// TODO: make configurable

// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md

func DeSerializationParameters() *iotago.DeSerializationParameters {
	return testdeserparams.DeSerializationParameters() // TODO
}

func RentStructure() *iotago.RentStructure {
	return testdeserparams.RentStructure() // TODO
}

const NetworkID uint64 = 0 // TODO
