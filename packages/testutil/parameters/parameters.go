package parameters

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/parameters"
)

var l1ParamsForTesting = &parameters.L1Params{
	// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Message fields without payload = max size of the payload
	MaxPayloadSize: parameters.MaxPayloadSize,
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
	BaseToken: &parameters.BaseToken{
		Name:            "Iota",
		TickerSymbol:    "MIOTA",
		Unit:            "MIOTA",
		Subunit:         "IOTA",
		Decimals:        6,
		UseMetricPrefix: false,
	},
}

func GetL1ParamsForTesting() *parameters.L1Params {
	return l1ParamsForTesting
}

func GetL1ProtocolParamsForTesting() *iotago.ProtocolParameters {
	return l1ParamsForTesting.Protocol
}

func GetBech32HRP() iotago.NetworkPrefix {
	return l1ParamsForTesting.Protocol.Bech32HRP
}
