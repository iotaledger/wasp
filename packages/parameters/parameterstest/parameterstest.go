// Package parameterstest provides testing utilities for the parameters package.
package parameterstest

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

var L1Mock = &parameters.L1Params{
	Protocol: &parameters.Protocol{
		Epoch:                 iotajsonrpc.NewBigInt(100),
		ProtocolVersion:       iotajsonrpc.NewBigInt(1),
		SystemStateVersion:    iotajsonrpc.NewBigInt(1),
		ReferenceGasPrice:     iotajsonrpc.NewBigInt(1000),
		EpochStartTimestampMs: iotajsonrpc.NewBigInt(1734538812318),
		EpochDurationMs:       iotajsonrpc.NewBigInt(86400000),
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
