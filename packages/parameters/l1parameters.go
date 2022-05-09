package parameters

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

// L1 describes parameters coming from the L1 node
type L1 struct {
	MaxTransactionSize int
	Protocol           *iotago.ProtocolParameters
}

func (l1 *L1) RentStructure() *iotago.RentStructure {
	return &l1.Protocol.RentStructure
}

func L1ForTesting() *L1 {
	return &L1{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Message fields without payload = max size of the payload
		MaxTransactionSize: 32000,
		// https://github.com/muXxer/protocol-rfcs/blob/master/text/0032-dust-protection/0032-dust-protection.md
		Protocol: &iotago.ProtocolParameters{
			Version:     tpkg.TestProtoParas.Version,
			NetworkName: tpkg.TestProtoParas.NetworkName,
			Bech32HRP:   tpkg.TestProtoParas.Bech32HRP,
			MinPoWScore: tpkg.TestProtoParas.MinPoWScore,
			RentStructure: iotago.RentStructure{
				VByteCost:    10,
				VBFactorData: 1,
				VBFactorKey:  1,
			},
			TokenSupply: tpkg.TestProtoParas.TokenSupply,
		},
	}
}
