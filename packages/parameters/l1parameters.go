package parameters

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"
)

// Parameters needed for interaction with layer1

// TODO  - get the max tx size from iotago somehow
// const MaxTransactionSize = iotago.Transaction.

// L1 describes parameters coming from the L1 node
type L1 struct {
	NetworkName               string
	NetworkID                 uint64
	Bech32Prefix              iotago.NetworkPrefix
	MaxTransactionSize        int
	DeSerializationParameters *iotago.DeSerializationParameters
}

func (l1 *L1) RentStructure() *iotago.RentStructure {
	return l1.DeSerializationParameters.RentStructure
}

func L1ForTesting() *L1 {
	return &L1{
		NetworkName:        "iota",
		Bech32Prefix:       "iota",
		NetworkID:          0,
		MaxTransactionSize: 32000,
		// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md
		DeSerializationParameters: testdeserparams.DeSerializationParameters(),
	}
}
