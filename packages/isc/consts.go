package isc

import "github.com/iotaledger/wasp/clients/iota-go/iotaclient"

const (
	Million         = 1_000_000
	GasCoinMaxValue = 1 * Million
)

// TODO Add the comprehensive top up calculation logic, then we can remvoe this constant

const TopUpFeeMin = iotaclient.DefaultGasBudget * 5
