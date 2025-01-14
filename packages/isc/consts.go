package isc

import "github.com/iotaledger/wasp/clients/iota-go/iotaclient"

const (
	Million         = 1_000_000
	GasCoinMaxValue = 1 * Million
)

// This threshold defines the amount of funds, which must always be enough to cover
// the gas costs of chain state transition for any number requests within supported range.
const maxTotalGasCostForChainStateTransition = iotaclient.DefaultGasBudget * 5

// TODO Add the comprehensive top up calculation logic, then we can remove this constant
const GasCoinMinBalance = maxTotalGasCostForChainStateTransition
