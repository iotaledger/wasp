// Package parameterstest provides testing utilities for the parameters package.
package parameterstest

import (
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

var L1Mock = &parameters.L1Params{
	Protocol: &parameters.Protocol{
		Epoch:                 iotagraphql.NewBigInt(100),
		ProtocolVersion:       iotagraphql.NewBigInt(1),
		SystemStateVersion:    iotagraphql.NewBigInt(1),
		ReferenceGasPrice:     iotagraphql.NewBigInt(1000),
		EpochStartTimestampMs: iotagraphql.NewBigInt(1734538812318),
		EpochDurationMs:       iotagraphql.NewBigInt(86400000),
	},
	BaseToken: &parameters.IotaCoinInfo{
		CoinType:    coin.BaseTokenType,
		Name:        "Iota",
		Symbol:      "IOTA",
		Description: "IOTA",
		IconURL:     "http://iota.org",
		Decimals:    parameters.BaseTokenDecimals,
		TotalSupply: 9978371123948460000,
	},
}
