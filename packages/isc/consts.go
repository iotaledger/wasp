package isc

import "github.com/iotaledger/wasp/v2/clients/iotagraphql"

const (
	Million = 1_000_000
)

// GasCoinTargetValue is the target value for topping up the gas coin. After
// each VM run, the gas coin will be topped up taking funds from the common
// account.
const GasCoinTargetValue = iotagraphql.DefaultGasBudget * 5
